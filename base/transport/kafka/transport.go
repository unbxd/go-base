package kafka

type (
	// TransportOption is set of options supported by the Transport
	TransportOption func(*Transport)

	// Transport is Kafka based transport
	Transport struct{}
)

// NewTransport returns the a new transport
func NewTransport() (*Transport, error) {
	return nil, nil
}
