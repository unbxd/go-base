package zook

import (
	"strings"
	"time"

	"github.com/samuel/go-zookeeper/zk"
	"github.com/unbxd/go-base/data/driver"
	"github.com/unbxd/go-base/errors"
)

type (
	// Driver defines zookeeper driver for Albus
	Driver struct {
		root    string
		timeout time.Duration
		acl     []zk.ACL

		servers []string

		conn *zk.Conn
	}

	DriverOption func(*Driver)
)

func check(conn *zk.Conn, root string) error {
	_, _, err := conn.Get(root)
	switch {
	case err != nil && (err == zk.ErrInvalidPath || err == zk.ErrNoNode):
		if _, er := conn.Create(
			root,
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
func (d *Driver) Open() error {
	// TODO: event channel, check what it does
	conn, _, err := zk.Connect(d.servers, d.timeout)
	if err != nil {
		return errors.Wrap(err, "Error initializing ZK Driver")
	}

	d.conn = conn
	d.acl = zk.WorldACL(zk.PermAll)

	return check(d.conn, d.root)
}

func (d *Driver) makePath(path string) error {
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
func (d *Driver) Read(path string) ([]byte, error) {
	data, _, err := d.conn.Get(path)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// Write writes the content to the path
func (d *Driver) Write(path string, data []byte) error {
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
func (d *Driver) Children(path string) ([]string, error) {

	// TODO: _ is acctually an event, see what it does
	children, _, err := d.conn.Children(path)
	if err != nil {
		return nil, err
	}

	return children, nil
}

// Delete deletes the node and all its children
func (d *Driver) Delete(path string) error {
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
func (d *Driver) Watch(path string) ([]byte, <-chan *driver.Event, error) {
	var channel = make(chan *driver.Event)

	val, _, ech, err := d.conn.GetW(path)
	if err != nil {
		return nil, nil, err
	}

	go func(path string, channel chan *driver.Event) {
		for {
			event := <-ech

			val, _, ech, err = d.conn.GetW(path)
			if err != nil {
				close(channel)
				return
			}

			switch event.Type {
			case zk.EventNodeCreated:
				channel <- &driver.Event{Type: driver.EventCreated, P: path, D: val, Err: err}
			case zk.EventNodeDeleted:
				channel <- &driver.Event{Type: driver.EventDeleted, P: path, D: val, Err: err}
			case zk.EventNodeDataChanged:
				channel <- &driver.Event{Type: driver.EventDataChanged, P: path, D: val, Err: err}
			case zk.EventNodeChildrenChanged:
				channel <- &driver.Event{Type: driver.EventChildrenChanged, P: path, D: val, Err: err}
			}

		}
	}(path, channel)

	return val, channel, nil
}

func (d *Driver) WatchChildren(path string) ([]string, <-chan *driver.Event, error) {
	var channel = make(chan *driver.Event)

	val, _, ech, err := d.conn.ChildrenW(path)
	if err != nil {
		return nil, nil, err
	}

	go func(path string, channel chan *driver.Event) {
		for {
			event := <-ech

			val, _, ech, err = d.conn.ChildrenW(path)
			if err != nil {
				close(channel)
				return
			}

			// This is done to wrap Zookeeper Events into Driver Events
			// This will ensure the re-usability of the interface
			switch event.Type {
			case zk.EventNodeCreated:
				channel <- &driver.Event{Type: driver.EventCreated, P: path, D: val}
			case zk.EventNodeDeleted:
				channel <- &driver.Event{Type: driver.EventDeleted, P: path, D: val}
			case zk.EventNodeDataChanged:
				channel <- &driver.Event{Type: driver.EventDataChanged, P: path, D: val}
			case zk.EventNodeChildrenChanged: //we will only get this event
				channel <- &driver.Event{Type: driver.EventChildrenChanged, P: path, D: val}
			}
		}
	}(path, channel)

	return val, channel, nil
}

// Close shuts down connection for the driver
func (d *Driver) Close() error {
	d.conn.Close()
	return nil
}

func (d *Driver) IsConnected() bool {
	state := d.conn.State()
	return state == zk.StateConnected || state == zk.StateHasSession
}

func (d *Driver) State() zk.State {
	return d.conn.State()
}

func WithACL(acl []zk.ACL) DriverOption {
	return func(d *Driver) {
		d.acl = acl
	}
}

func WithTimeout(timeout time.Duration) DriverOption {
	return func(d *Driver) {
		d.timeout = timeout
	}
}

func WithRootDirectory(root string) DriverOption {
	return func(d *Driver) {
		d.root = root
	}
}

// NewZKDriver returns new zookeeper driver
func NewZKDriver(servers []string, options ...DriverOption) driver.Driver {
	driver := &Driver{
		servers: servers,
		timeout: 18 * time.Second,
		root:    "/",
		acl:     zk.WorldACL(zk.PermAll),
	}

	for _, fn := range options {
		fn(driver)
	}

	return driver
}
