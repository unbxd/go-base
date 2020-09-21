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
	// Watch watches for changes on the node
	Watch(path string) (<-chan Event, <-chan error)
	// Watch tree for a change
	WatchTree(path string, level int) (<-chan Event, <-chan error)
	// Close shuts down the connection for the driver
	Close() error
}
