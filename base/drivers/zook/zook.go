package zook

import (
	"strings"
	"sync"
	"time"

	"github.com/unbxd/go-base/base/drivers"

	"github.com/pkg/errors"
	"github.com/samuel/go-zookeeper/zk"
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
func (d *ZookDriver) Watch(path string) ([]byte, <-chan *drivers.Event, error) {
	var channel = make(chan *drivers.Event)

	val, _, ech, err := d.conn.GetW(path)
	if err != nil {
		return nil, nil, err
	}

	go func(path string, channel chan *drivers.Event) {

		for {
			select {
			case event := <-ech:
				val, _, ech, err = d.conn.GetW(path)

				switch event.Type {
				case zk.EventNodeCreated:
					channel <- &drivers.Event{Type: drivers.EventCreated, P: path, D: val, Err: err}
				case zk.EventNodeDeleted:
					channel <- &drivers.Event{Type: drivers.EventDeleted, P: path, D: val, Err: err}
				case zk.EventNodeDataChanged:
					channel <- &drivers.Event{Type: drivers.EventDataChanged, P: path, D: val, Err: err}
				case zk.EventNodeChildrenChanged:
					channel <- &drivers.Event{Type: drivers.EventChildrenChanged, P: path, D: val, Err: err}
				}

				if err != nil {
					close(channel)
					return
				}
			}
		}
	}(path, channel)

	return val, channel, nil
}

// WatchChildren watches for changes on node and its children
func (d *ZookDriver) WatchChildren(path string) ([]string, <-chan *drivers.Event, error) {
	var channel = make(chan *drivers.Event)
	val, _, ech, err := d.conn.ChildrenW(path)
	if err != nil {
		return nil, nil, err
	}
	//add to zook driver, one entry per children
	chCh := make([]<-chan zk.Event, len(val))
	for i := 0; i < len(val); i++ {
		_, _, ech, err := d.conn.GetW(path + "/" + val[i])
		if err != nil {
			close(channel)
		}
		chCh[i] = ech
	}
	//adding parent event channel too
	chCh = append(chCh, ech)
	//merging all events into one channel
	mCh := merge(chCh)
	go func(path string, channel chan *drivers.Event) {

		for {
			select {
			case event := <-ech:
				val, _, ech, err = d.conn.ChildrenW(path)
				//add watch on this new node
				// This is done to wrap Zookeeper Events into Driver Events
				// This will ensure the re-usability of the interface
				switch event.Type {
				case zk.EventNodeCreated:
					newC, _, uCh, err := d.conn.GetW(event.Path)
					if err != nil {
						close(channel)
					}
					tmp := make([]<-chan zk.Event, 1)
					tmp = append(tmp, uCh)
					tmp = append(tmp, mCh)
					mCh = merge(tmp)
					channel <- &drivers.Event{Type: drivers.EventCreated, P: path, D: string(newC)}
				case zk.EventNodeDeleted:
					delC, _, uCh, err := d.conn.GetW(event.Path)
					if err != nil {
						close(channel)
					}
					tmp := make([]<-chan zk.Event, 1)
					tmp = append(tmp, uCh)
					tmp = append(tmp, mCh)
					mCh = merge(tmp)
					channel <- &drivers.Event{Type: drivers.EventDeleted, P: path, D: string(delC)}
				case zk.EventNodeDataChanged:
					chC, _, uCh, err := d.conn.GetW(event.Path)
					if err != nil {
						close(channel)
					}
					tmp := make([]<-chan zk.Event, 1)
					tmp = append(tmp, uCh)
					tmp = append(tmp, mCh)
					mCh = merge(tmp)
					channel <- &drivers.Event{Type: drivers.EventDataChanged, P: path, D: string(chC)}
				case zk.EventNodeChildrenChanged:
					if event.Path == path {
						//find new added child and add its channel to mECh
						nChList, _, err := d.conn.Children(path)
						if err != nil {
							close(channel)
						}
						diffChList := difference(nChList, val)
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
					channel <- &drivers.Event{Type: drivers.EventChildrenChanged, P: path, D: val}
				}

				if err != nil {
					close(channel)
					return
				}
			}
		}
	}(path, channel)

	return val, channel, nil
}

// Close shuts down connection for the driver
func (d *ZookDriver) Close() error {
	d.conn.Close()
	return nil
}

func (d *ZookDriver) IsConnected() bool {
	state := d.conn.State()
	return state == zk.StateConnected || state == zk.StateHasSession
}

func (d *ZookDriver) State() zk.State {
	return d.conn.State()
}

// NewZKDriver returns new zookeeper driver
func NewZKDriver(servers []string, timeout time.Duration, rootDir string) drivers.Driver {
	return &ZookDriver{
		servers: servers,
		timeout: timeout,
		rootDir: rootDir,
	}
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
