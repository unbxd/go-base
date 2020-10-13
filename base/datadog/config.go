package datadog

import (
	"github.com/DataDog/datadog-go/statsd"
	"github.com/pkg/errors"
)

// Config defines the Datadog Configuration
type Config struct {
	URL        string
	Namespace  string
	SkipErrors bool
	Tags       []string
}

// Build builds the client based on configuration
func (c *Config) Build() (*Client, error) {
	var (
		enabled bool
		err     error
	)

	stc, err := statsd.New(c.URL)
	if err != nil {
		return nil, errors.Wrap(err, "[ut.dd] create client failed")
	}

	stc.Namespace = c.Namespace
	stc.Tags = c.Tags
	stc.SkipErrors = c.SkipErrors

	return &Client{stc, enabled}, nil
}
