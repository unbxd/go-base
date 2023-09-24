package cb

import (
	"context"
	"strings"
	"sync"

	"github.com/DataDog/datadog-go/statsd"
	"github.com/unbxd/go-base/endpoint"
	"github.com/unbxd/go-base/log"
	"github.com/unbxd/go-base/metrics"
	cbplugins "github.com/unbxd/go-base/net/cb/plugins"

	"github.com/pkg/errors"
	"github.com/unbxd/hystrix-go/hystrix"
	"github.com/unbxd/hystrix-go/hystrix/metric"
	"github.com/unbxd/hystrix-go/plugins"
)

type (
	BreakerConf struct {
		Enable       bool
		Timeout      int
		MaxConc      int
		VolThrs      int
		SlpWind      int
		ErrPerctThrs int
		Prefix       string
	}

	configured struct {
		in map[string]struct{}
		mu sync.Mutex
	}

	BreakerAfterFunc func(req interface{}, res interface{}, err error)

	// Breaker wraps the endpointer and the command
	// config required for the hysterix
	Breaker struct {
		enable     bool
		cmdcfg     *hystrix.CommandConfig
		fn         endpoint.Endpoint
		fallbackfn func(error) error
		cfgred     *configured
		cmdPrefix  string
		afterFunc  BreakerAfterFunc
	}

	// BreakerOption is options that modify the Breaker
	BreakerOption func(*Breaker) error

	// Commander is interface implemented for breaker to
	// extract the command from a method
	Commander interface {
		Command() string
	}
)

func (cf *configured) Has(cmd string) bool {
	cf.mu.Lock()
	defer cf.mu.Unlock()

	_, ok := cf.in[cmd]
	return ok
}

func (cf *configured) Add(cmd string) {
	cf.mu.Lock()
	defer cf.mu.Unlock()

	cf.in[cmd] = struct{}{}
}

func (b *Breaker) command(rqi interface{}) (string, error) {
	req, ok := rqi.(Commander)
	if !ok {
		// we should never reach here
		return "", errors.New("request is not http.Request")
	}

	var (
		buf strings.Builder
	)

	if b.cmdPrefix != "" {
		buf.WriteString(b.cmdPrefix)
		buf.WriteRune('-')
		buf.WriteString(req.Command())
	} else {
		buf.WriteString(req.Command())
	}

	return buf.String(), nil
}

// Endpoint returns an endpoint which has circuit breaker
// wraped around it
func (b *Breaker) Endpoint() endpoint.Endpoint {
	return func(
		cx context.Context,
		rqi interface{},
	) (rsi interface{}, err error) {
		if !b.enable {
			return b.fn(cx, rqi)
		}

		_, ok := rqi.(Commander)
		if !ok {
			// do nothing, use b.fn
			return b.fn(cx, rqi)
		}

		cmd, err := b.command(rqi)
		if err != nil {
			return b.fn(cx, rqi)
		}

		// check if there is a config for
		// command name in hysterix
		// _, ok = b.configured[]
		// if !ok {
		// 	hystrix.ConfigureCommand(cmd, *b.cmdcfg)
		// }

		if !b.cfgred.Has(cmd) {
			hystrix.ConfigureCommand(
				cmd, *b.cmdcfg,
			)
			b.cfgred.Add(cmd)
		}

		rc := make(chan interface{}, 1)
		ec := hystrix.Go(cmd, func() (er error) {
			res, er := b.fn(cx, rqi)
			if er != nil {
				return er
			}

			rc <- res
			return
		}, b.fallbackfn)

		select {
		case rsi = <-rc:
			break
		case err = <-ec:
			break
		}
		b.afterFunc(rqi, rsi, err)
		return
	}
}

// NewBreaker returns a circuit breaker
func NewBreaker(fn endpoint.Endpoint, opts ...BreakerOption) (*Breaker, error) {
	bk := &Breaker{
		fn: fn,
		cmdcfg: &hystrix.CommandConfig{
			// DefaultTimeout is how long to wait for command to
			// complete, in millisecond
			Timeout: 30000,
			// DefaultMaxConcurrent is how many commands of the
			// same type can run at the same time
			MaxConcurrentRequests: hystrix.DefaultMaxConcurrent,
			// DefaultVolumeThreshold is the minimum number of
			// requests needed before a circuit can be tripped due
			// to health
			RequestVolumeThreshold: hystrix.DefaultVolumeThreshold,
			// DefaultSleepWindow is how long, in milliseconds, to
			// wait after a circuit opens before testing for
			// recovery
			SleepWindow: hystrix.DefaultSleepWindow,
			// DefaultErrorPercentThreshold causes circuits to open
			// once the rolling measure of errors exceeds this
			// percent of requests
			ErrorPercentThreshold: hystrix.DefaultErrorPercentThreshold,
		},
		cfgred: &configured{
			in: make(map[string]struct{}),
		},
	}

	for _, o := range opts {
		if er := o(bk); er != nil {
			return nil, er
		}
	}

	return bk, nil
}

// WithCommandPrefix sets the prefix for the hysterix command
func WithCommandPrefix(prefix string) BreakerOption {
	return func(b *Breaker) (err error) {
		b.cmdPrefix = prefix
		return
	}
}

// WithTimeout returns a circuit breaker with a given timeout
func WithTimeout(inmils int) BreakerOption {
	return func(b *Breaker) (err error) {
		b.cmdcfg.Timeout = inmils
		return
	}
}

// WithMaxConcurrentRequests sets configuration for max concurrent
// requests supported for a given command
func WithMaxConcurrentRequests(count int) BreakerOption {
	return func(b *Breaker) (err error) {
		b.cmdcfg.MaxConcurrentRequests = count
		return
	}
}

// WithRequestVolumeThreshold sets the minimum number of request needed before
// the circuit can be tripped
func WithRequestVolumeThreshold(count int) BreakerOption {
	return func(b *Breaker) (err error) {
		b.cmdcfg.RequestVolumeThreshold = count
		return
	}
}

// WithSleepWindow sets the time in millisecond to wait after a circuit
// opens for testing for recovery
func WithSleepWindow(inmils int) BreakerOption {
	return func(b *Breaker) (err error) {
		b.cmdcfg.SleepWindow = inmils
		return
	}
}

// WithErrorPercentageThreshold the percentage threshold beyond which
// the circuit will be deemed open
func WithErrorPercentageThreshold(count int) BreakerOption {
	return func(b *Breaker) (err error) {
		b.cmdcfg.ErrorPercentThreshold = count
		return
	}
}

// WithDatadogClient sets custom datadog event emitter for
// Tolerance
func WithDatadogClient(client *statsd.Client) BreakerOption {
	return func(tp *Breaker) error {
		metric.Registry.Register(
			plugins.NewDatadogCollectorWithClient(client),
		)
		return nil
	}
}

// WithMetricsCollector sets the breaker with go-base metrics event emitter
func WithMetricsCollector(metrics metrics.Provider) BreakerOption {
	return func(tp *Breaker) error {
		metric.Registry.Register(
			cbplugins.NewMetricsCollector(metrics),
		)
		return nil
	}
}

// WithBreakerEnable sets if the breaker is enabled
func WithBreakerEnable(en bool) BreakerOption {
	return func(tp *Breaker) (err error) {
		tp.enable = en
		return
	}
}

func WithBreakerAfterFunc(b BreakerAfterFunc) BreakerOption {
	return func(tp *Breaker) (err error) {
		tp.afterFunc = b
		return
	}
}

func fnfn(v int, opts *[]BreakerOption, fn func(int) BreakerOption) {
	if v > 0 {
		*opts = append(*opts, fn(v))
	}
}

// NewBreakerFromConfig builds breaker from config
func NewBreakerFromConfig(
	fn endpoint.Endpoint, lg log.Logger, cfg *BreakerConf, opts ...BreakerOption,
) (*Breaker, error) {

	/*
		breaker:
		  timeout: 1000			# in millis
		  max_concurrent: 10
		  volume_threshold: 20
		  sleep_window: 5000		# in millis
		  error_percent_threshold: 50

	*/

	fnfn(cfg.Timeout, &opts, WithTimeout)
	fnfn(cfg.MaxConc, &opts, WithMaxConcurrentRequests)
	fnfn(cfg.VolThrs, &opts, WithRequestVolumeThreshold)
	fnfn(cfg.SlpWind, &opts, WithSleepWindow)
	fnfn(cfg.ErrPerctThrs, &opts, WithErrorPercentageThreshold)

	opts = append(
		opts,
		WithBreakerEnable(cfg.Enable),
		WithCommandPrefix(cfg.Prefix),
	)

	return NewBreaker(fn, opts...)
}
