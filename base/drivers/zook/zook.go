package zook

import (
	"strings"
	"sync"
	"time"

	"github.com/apoorvprecisely/go-base/base/drivers"

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
	var ch = make(chan *drivers.Event)
	return d.WatchOC(path, ch)
}

// WatchOC watches for changes on node with channel
func (d *ZookDriver) WatchOC(path string, channel chan *drivers.Event) ([]byte, <-chan *drivers.Event, error) {
	val, _, ech, err := d.conn.GetW(path)
	if err != nil {
		return nil, nil, err
	}

	go func(path string, channel chan *drivers.Event) {

		for {
			select {
			case event := <-ech:
				val, _, ech, err = d.conn.GetW(path)
				ch, _, errCh := d.conn.Children(path)
				if errCh != nil {
					ch = make([]string, 0)
				}
				switch event.Type {
				case zk.EventNodeCreated:
					channel <- &drivers.Event{Type: drivers.EventCreated, P: path, D: val, C: ch, Err: err}
				case zk.EventNodeDeleted:
					channel <- &drivers.Event{Type: drivers.EventDeleted, P: path, D: val, C: ch, Err: err}
				case zk.EventNodeDataChanged:
					channel <- &drivers.Event{Type: drivers.EventDataChanged, P: path, D: val, C: ch, Err: err}
				case zk.EventNodeChildrenChanged:
					channel <- &drivers.Event{Type: drivers.EventChildrenChanged, P: path, D: val, C: ch, Err: err}
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

// WatchDir watches for changes on node
func (d *ZookDriver) WatchDir(path string) ([]string, <-chan *drivers.Event, error) {
	var ch = make(chan *drivers.Event)
	go d.WatchDirOC(path, ch)
	return nil, ch, nil
}

// WatchDirOC watches for changes on node with channel
func (d *ZookDriver) WatchDirOC(path string, channel chan *drivers.Event) ([]string, <-chan *drivers.Event, error) {
	for {
		children, _, childEventCh, err := d.conn.ChildrenW(path)
		if err != nil {
			return nil, nil, err
		}
		for _, child := range children {
			go d.WatchDirOC(path+"/"+child, channel)

		}
		_, _, baseCh, err := d.conn.GetW(path)
		if err != nil {
			return nil, nil, err
		}
		chCh := make([]<-chan zk.Event, 2)
		chCh = append(chCh, childEventCh)
		chCh = append(chCh, baseCh)
		//merging all events into one channel
		mCh := merge(chCh)
		select {
		case e := <-mCh:
			val, _, err := d.conn.Get(path)
			ch, _, errCh := d.conn.Children(path)
			if errCh != nil {
				ch = make([]string, 0)
			}
			switch e.Type {
			case zk.EventNodeCreated:
				channel <- &drivers.Event{Type: drivers.EventCreated, P: e.Path, D: val, C: ch, Err: err}
			case zk.EventNodeDeleted:
				channel <- &drivers.Event{Type: drivers.EventDeleted, P: e.Path, D: val, C: ch, Err: err}
			case zk.EventNodeDataChanged:
				channel <- &drivers.Event{Type: drivers.EventDataChanged, P: e.Path, D: val, C: ch, Err: err}
			case zk.EventNodeChildrenChanged:
				newChildren, _, err := d.conn.Children(e.Path)
				if err != nil {
					return nil, nil, err
				}

				// a node was added -- watch the new node
				for _, i := range newChildren {
					if contains(children, i) {
						continue
					}

					newNode := e.Path + "/" + i
					// a new service was created under prefix
					go d.WatchDirOC(newNode, channel)

					nodes, _, _ := d.conn.Children(newNode)
					for _, node := range nodes {
						n := newNode + "/" + node
						go d.WatchOC(n, channel)
						val, _, err := d.conn.Get(n)
						if err != nil {
							continue
						}
						channel <- &drivers.Event{Type: drivers.EventChildrenChanged, P: path, D: val, Err: err}
					}

				}
			}
		}
	}
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
func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
