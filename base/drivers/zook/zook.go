package zook

import (
	"github.com/unbxd/go-base/base/drivers"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/samuel/go-zookeeper/zk"
	"sync"
)

// ZookDriver defines zookeeper driver for Albus
type ZookDriver struct {
	rootDir string
	servers []string
	timeout time.Duration
	conn    *zk.Conn
	acl     []zk.ACL
}

func (d *ZookDriver) sanitizeSetup(conn *zk.Conn) error {
	_, _, err := conn.Get(d.rootDir)
	switch {
	case err != nil && (err == zk.ErrInvalidPath || err == zk.ErrNoNode):
		if _, er := conn.Create(
			d.rootDir,
			[]byte("Emerge World!!"),
			int32(0),
			zk.WorldACL(zk.PermAll),
		); er != nil {
			return errors.Wrap(err, er.Error())
		}
	case err != nil && err != zk.ErrInvalidPath:
		return err
	}
	return nil
}

// Open initializes the driver
func (d *ZookDriver) Open() error {
	// TODO: event channel, check what it does
	conn, _, err := zk.Connect(d.servers, d.timeout)
	if err != nil {
		return errors.Wrap(err, "Error initializing ZK Driver")
	}

	err = d.sanitizeSetup(conn)
	if err != nil {
		return errors.Wrap(err, "Error stanitizing the setup")
	}

	d.conn = conn
	d.acl = zk.WorldACL(zk.PermAll)
	// return d.sanitizeSetup()
	return nil
}

func (d *ZookDriver) makePath(path string) error {
	pathSlice := strings.Split(path, "/")

	var cPath = ""
	for _, ele := range pathSlice[1:] {
		cPath = cPath + "/" + ele

		exists, _, err := d.conn.Exists(cPath)
		if err != nil {
			return errors.Wrap(err, "Error walking path")
		}

		if !exists {
			_, err := d.conn.Create(
				cPath,
				[]byte("{}"),
				int32(0),
				d.acl,
			)
			if err != nil {
				return errors.Wrap(err, "Error adding Node: ")
			}
		}
	}
	return nil
}

// Read reads the content from the path and returns the value in bytes
func (d *ZookDriver) Read(path string) ([]byte, error) {
	data, _, err := d.conn.Get(path)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// Write writes the content to the path
func (d *ZookDriver) Write(path string, data []byte) error {
	_, stat, err := d.conn.Get(path)
	if err != nil && err == zk.ErrNoNode {
		err := d.makePath(path)
		if err != nil {
			return err
		}
	}
	if err != nil && err != zk.ErrNoNode {
		return err
	}

	_, er := d.conn.Set(path, data, stat.Version)
	if er != nil {
		return errors.Wrap(err, "Error writing data to node. Path: "+path)
	}
	return nil
}

// Children returns the children of the path
func (d *ZookDriver) Children(path string) ([]string, error) {
	// TODO: _ is acctually an event, see what it does
	children, _, err := d.conn.Children(path)
	if err != nil {
		return nil, err
	}

	return children, nil
}

// Delete deletes the node and all its children
func (d *ZookDriver) Delete(path string) error {
	children, _, err := d.conn.Children(path)
	if err != nil {
		return err
	}
	if len(children) > 0 {
		for _, child := range children {
			err := d.Delete(path + "/" + child)
			if err != nil {
				return err
			}
		}
	} else {
		err := d.conn.Delete(path, -1)
		if err != nil {
			return err
		}
	}
	return nil
}

// Watch watches for changes on node
func (d *ZookDriver) Watch(path string) (<-chan drivers.Event, <-chan error) {
	var channel = make(chan drivers.Event)
	var errChannel = make(chan error)

	go func(path string, channel chan drivers.Event, errChannel chan error) {
		for {
			_, _, ech, err := d.conn.GetW(path)
			if err != nil {
				close(channel)
				errChannel <- err
				return
			}
			select {
			case event := <-ech:
				// This is done to wrap Zookeeper Events into Driver Events
				// This will ensure the re-usability of the interface
				switch event.Type {
				case zk.EventNodeCreated:
					channel <- drivers.Event{Type: drivers.EventCreated, P: path}
				case zk.EventNodeDeleted:
					channel <- drivers.Event{drivers.EventDeleted, path}
				case zk.EventNodeDataChanged:
					channel <- drivers.Event{drivers.EventDataChanged, path}
				case zk.EventNodeChildrenChanged:
					channel <- drivers.Event{drivers.EventChildrenChanged, path}
				}
			}
		}
	}(path, channel, errChannel)

	return channel, errChannel
}

func (d *ZookDriver) WatchTree(path string, level int) (<-chan drivers.Event, <-chan error) {
	//todo: assumes one level for now
	var channel = make(chan drivers.Event)
	var errChannel = make(chan error)
	chList, _, rnCh, err := d.conn.ChildrenW(path)
	if err != nil {
		close(channel)
	}
	//add to zook driver, one entry per children
	chCh := make([]<-chan zk.Event, len(chList))
	for i := 0; i < len(chList); i++ {
		_, _, ech, err := d.conn.GetW(path + "/" + chList[i])
		if err != nil {
			close(channel)
		}
		chCh[i] = ech
	}
	//adding parent event channel too
	chCh = append(chCh, rnCh)
	//merging all events into one channel
	mCh := merge(chCh)
	go func(path string, mCh <-chan zk.Event, channel chan drivers.Event, errChannel chan error) {
		for {
			select {
			case event := <-mCh:
				// This is done to wrap Zookeeper Events into Driver Events
				// This will ensure the re-usability of the interface
				switch event.Type {
				case zk.EventNodeCreated:
					_, _, uCh, err := d.conn.GetW(event.Path)
					if err != nil {
						close(channel)
					}
					tmp := make([]<-chan zk.Event, 1)
					tmp = append(tmp, uCh)
					tmp = append(tmp, mCh)
					mCh = merge(tmp)
					channel <- drivers.Event{drivers.EventCreated, event.Path}

				case zk.EventNodeDeleted:
					_, _, uCh, err := d.conn.GetW(event.Path)
					if err != nil {
						close(channel)
					}
					tmp := make([]<-chan zk.Event, 1)
					tmp = append(tmp, uCh)
					tmp = append(tmp, mCh)
					mCh = merge(tmp)
					channel <- drivers.Event{drivers.EventDeleted, event.Path}

				case zk.EventNodeDataChanged:
					_, _, uCh, err := d.conn.GetW(event.Path)
					if err != nil {
						close(channel)
					}
					tmp := make([]<-chan zk.Event, 1)
					tmp = append(tmp, uCh)
					tmp = append(tmp, mCh)
					mCh = merge(tmp)
					channel <- drivers.Event{drivers.EventDataChanged, event.Path}

				case zk.EventNodeChildrenChanged:
					if event.Path == path {
						//find new added child and add its channel to mECh
						nChList, _, err := d.conn.Children(path)
						if err != nil {
							close(channel)
						}
						diffChList := difference(nChList, chList)
						tmp := make([]<-chan zk.Event, len(diffChList))
						for _, child := range diffChList {
							_, _, ech, err := d.conn.GetW(path + "/" + child)
							if err != nil {
								close(channel)
							}
							tmp = append(tmp, ech)
						}
						_, _, rCh, err := d.conn.ChildrenW(path)
						if err != nil {
							close(channel)
						}
						tmp = append(tmp, rCh)
						tmp = append(tmp, mCh)
						mCh = merge(tmp)
					}
				}
			}
		}
	}(path, mCh, channel, errChannel)

	return channel, errChannel
}
func difference(slice1 []string, slice2 []string) []string {
	diffStr := []string{}
	m := map[string]int{}

	for _, s1Val := range slice1 {
		m[s1Val] = 1
	}
	for _, s2Val := range slice2 {
		m[s2Val] = m[s2Val] + 1
	}

	for mKey, mVal := range m {
		if mVal == 1 {
			diffStr = append(diffStr, mKey)
		}
	}

	return diffStr
}
func merge(cs []<-chan zk.Event) <-chan zk.Event {
	var wg sync.WaitGroup
	out := make(chan zk.Event)

	output := func(c <-chan zk.Event) {
		for n := range c {
			out <- n
		}
		wg.Done()
	}
	wg.Add(len(cs))
	for _, c := range cs {
		go output(c)
	}

	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}

// Close shuts down connection for the driver
func (d *ZookDriver) Close() error {
	d.conn.Close()
	return nil
}

// NewZKDriver returns new zookeeper driver
func NewZKDriver(servers []string, timeout time.Duration) drivers.Driver {
	return &ZookDriver{
		servers: servers,
		timeout: timeout,
	}
}
