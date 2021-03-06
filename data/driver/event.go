package driver

type EventType uint64

func (e EventType) String() string {
	str, ok := eventName[e]
	if !ok {
		return "UNKNOWN"
	}
	return str
}

// Driver Events
const (
	EventCreated         EventType = 1
	EventDeleted         EventType = 2
	EventDataChanged     EventType = 3
	EventChildrenChanged EventType = 4
)

var eventName = map[EventType]string{
	EventCreated:         "Created",
	EventDeleted:         "Deleted",
	EventDataChanged:     "Data Changed",
	EventChildrenChanged: "Children Changed",
}

// Event is defined to wrap events emitted by the driver
type Event struct {
	Type EventType
	P    string
	D    interface{}
	Err  error
}

// EventType returns event type of the event
func (e *Event) EventType() EventType { return e.Type }

// Path returns the path of event
func (e *Event) Path() string { return e.P }

func (e *Event) Data() interface{} { return e.D }

func (e *Event) Error() error { return e.Err }
