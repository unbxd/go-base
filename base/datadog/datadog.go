package datadog

import (
	"github.com/DataDog/datadog-go/statsd"
	"github.com/pkg/errors"
)

// Client is wrapper on top of statsd.Client
type Client struct {
	*statsd.Client

	enabled bool
}

// Enabled returns if the datadog client is enabled
func (c *Client) Enabled() bool { return c.enabled }

// NewClient returns new DataDog client
func NewClient(cfg *Config) (*Client, error) {
	var (
		client *Client

		err error
	)

	client, err = cfg.Build()
	if err != nil {
		return nil, errors.Wrap(err, "[util.dd] create client failed")
	}

	return client, err
}
