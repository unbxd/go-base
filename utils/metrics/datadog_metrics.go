package metrics

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	kitlogger "github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics/dogstatsd"
	"github.com/mitchellh/mapstructure"
	"github.com/unbxd/go-base/utils/log"
)

type (
	// datadog is wrapper on top of statsd.Client
	datadog struct {
		dstd *dogstatsd.Dogstatsd

		connstr string
		host    string
		port    string

		ns  string
		lvs []string

		tick   time.Duration
		logger kitlogger.Logger

		enabled bool
	}

	// DatadogOption provides way to modify the client object
	DatadogOption func(*datadog)
)

// WithDatadogNamespace sets the namespace for Datadog Metrics
func WithDatadogNamespace(ns string) DatadogOption {
	return func(dd *datadog) { dd.ns = ns }
}

// WithDatadogServerHost sets the server URI for datadog to connect to
func WithDatadogServerHost(host string) DatadogOption {
	return func(dd *datadog) { dd.host = host }
}

// WithDatadogServerPort sets the server URI for datadog to connect to
func WithDatadogServerPort(port string) DatadogOption {
	return func(dd *datadog) { dd.port = port }
}

// WithDatadogServerConnstr sets the server URI for datadog to connect to
func WithDatadogServerConnstr(cstr string) DatadogOption {
	return func(dd *datadog) { dd.connstr = cstr }
}

// WithDatadogTag appends a tag for existing set of tags
func WithDatadogTag(key, value string) DatadogOption {
	return func(dd *datadog) { dd.lvs = append(dd.lvs, []string{key, value}...) }
}

// WithDatadogLabelValues overwrites the existing tags with a new  set of tags provided
// Note, that LabelValues are pair of key and a value past as consecutive elements
// in the arrray. For instance, if key and value pair are `go-base:true`, pass it
// as []string{"go-base", "true"}
// Internally the count for labelvalues is checked and if found odd, it panics
func WithDatadogLabelValues(lvs []string) DatadogOption {
	return func(dd *datadog) { dd.lvs = lvs }
}

// WithDatadogTags is wrapper on label values where the expected format for
// tags is traditional `key:value` pair. If token doesn't have a `:` separator
// it is ignored.
func WithDatadogTags(tags []string) DatadogOption {
	var lvs []string

	for _, tag := range tags {
		ss := strings.Split(tag, ":")
		if len(ss) == 2 {
			lvs = append(lvs, []string{ss[0], ss[1]}...)
		}
	}

	return func(dd *datadog) { dd.lvs = append(dd.lvs, lvs...) }
}

// WithDatadogEnabled toggles the datadog metrics on or off
func WithDatadogEnabled(enabled bool) DatadogOption {
	return func(dd *datadog) { dd.enabled = enabled }
}

// WithDatadogTickInSeconds sets the send loop timer in seconds
func WithDatadogTickInSeconds(tick int) DatadogOption {
	return func(dd *datadog) { dd.tick = time.Second * time.Duration(tick) }
}

// WithDatadogLogger sets the logger for Datadog Stats
func WithDatadogLogger(logger log.Logger) DatadogOption {
	return func(dd *datadog) { dd.logger = logger }
}

// WithDatadogConfigObject sets the properties for datadog using
// older config object used in past projects. This function is exclusively
// for backward compatibility.
// Here is the expected configuration when read in YAML
//
// 	url: "datadog:8125"
// 	namespace: "wingman"
// 	tags:
// 	  - "env:compose"
// 	  - "ci:true"
//
// the structure which reads the config is defined as
//
// 	type Config struct {
// 		URL        string `json:"url" yaml:"url"`
// 		Namespace  string `json:"namespace" yaml:"namespace"`
// 		Tags       []string `json:"tags" yaml:"tags"`
// 	}
// Note: skipErrors is not required anymore
func WithDatadogConfigObject(cfg interface{}) DatadogOption {
	type config struct {
		URL       string   `mapstructure:"url"`
		Namespace string   `mapstructure:"namespace"`
		Tags      []string `mapstructure:"tags"`
	}

	var (
		cf  config
		err error
	)

	err = mapstructure.Decode(cfg, &cf)
	if err != nil {
		// panic because of programmer error
		panic(
			fmt.Sprintf("programmer error: cfg is not config. err: [%s]", err.Error()),
		)
	}

	return func(dd *datadog) {
		dd.connstr = cf.URL
		dd.ns = cf.Namespace
		WithDatadogTags(cf.Tags)(dd)
	}

}

func (dd *datadog) NewCounter(
	name string, sampleRate float64,
) Counter {
	return Counter(
		dd.dstd.NewCounter(name, sampleRate),
	)
}

func (dd *datadog) NewHistogram(
	name string, sampleRate float64,
) Histogram {
	return Histogram(
		dd.dstd.NewHistogram(name, sampleRate),
	)
}

func (dd *datadog) NewGauge(name string) Gauge {
	return Gauge(dd.dstd.NewGauge(name))
}

// NewDatadogMetrics returns metrics which supports Datadog
func NewDatadogMetrics(opts ...DatadogOption) (Metrics, error) {
	dd := &datadog{
		connstr: "",
		host:    "localhost",
		port:    "8125",
		ns:      "gb",
		tick:    10 * time.Second,
		enabled: true,
		logger:  kitlogger.NewNopLogger(),
	}

	for _, o := range opts {
		o(dd)
	}

	if dd.connstr == "" {
		// build from host port
		dd.connstr = net.JoinHostPort(dd.host, dd.port)
	}

	dd.dstd = dogstatsd.New(
		dd.ns, dd.logger, dd.lvs...,
	)

	go func() {
		//nolint:errcheck
		dd.logger.Log("[metrics/dd]",
			"starting backgound sendloop",
			"address", dd.connstr,
		)
		dd.dstd.SendLoop(
			context.Background(),
			time.Tick(dd.tick),
			"udp",
			dd.connstr,
		)
	}()

	return dd, nil
}
