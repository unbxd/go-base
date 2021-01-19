package drivers

// Driver defines an interface for reading/writing data
// to data source
type Driver interface {
	// Open opens the driver
	Open() error
	// Read reads the node from unix like tree path
	Read(path string) ([]byte, error)
	// Write writes to designated location
	Write(path string, data []byte) error
	// Children returns the list of children for a path
	Children(path string) ([]string, error)
	// Delete deletes the node in path
	Delete(path string) error
	// Watch gets the value and watches for future changes
	Watch(path string) ([]byte, <-chan *Event, error)
	// Watch gets the children and watches for future changes
	WatchDir(path string) ([]string, <-chan *Event, error)
	// Close shuts down the connection for the driver
	Close() error
}
