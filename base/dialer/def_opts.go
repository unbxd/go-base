package dialer

import (
	"context"
	"net"
	"net/http"
	"time"

	"github.com/pkg/errors"
)

type (
	Conf struct {
		Tr *TransportConf
		Nw *NetworkConf
		To *TimeoutConf
	}

	TransportConf struct {
		// MaxIdleConns controls the maximum number of idle (keep-alive)
		// connections across all hosts. Zero means no limit.
		MaxIdleConns int

		// MaxIdleConnsPerHost, if non-zero, controls the maximum idle
		// (keep-alive) connections to keep per-host.
		MaxIdleConnsPerHost int

		// MaxConnsPerHost optionally limits the total number of
		// connections per host, including connections in the dialing,
		// active, and idle states. On limit violation, dials will block.
		//
		// Zero means no limit.
		MaxConnsPerHost int

		// IdleConnTimeout is the maximum amount of time an idle
		// IdleConnTimeout is the maximum amount of time an idle
		// itself.
		// Zero means no limit.
		IdleConnTimeout int
	}

	NetworkConf struct {
		// Timeout is the maximum amount of time a dial will wait for
		// a connect to complete. If Deadline is also set, it may fail
		// earlier.
		//
		// The default is no timeout.
		//
		// When using TCP and dialing a host name with multiple IP
		// addresses, the timeout may be divided between them.
		//
		// With or without a timeout, the operating system may impose
		// its own earlier timeout. For instance, TCP timeouts are
		// often around 3 minutes.
		Timeout int

		// KeepAlive specifies the keep-alive period for an active
		// network connection.
		// If zero, keep-alives are enabled if supported by the protocol
		// and operating system. Network protocols or operating systems
		// that do not support keep-alives ignore this field.
		// If negative, keep-alives are disabled.
		KeepAlive int
	}

	TimeoutConf struct {
		Tm int //in millisecond
	}
)

// WithRoundTripperExecutor executor which uses a custom round tripper built
// from configurations to call downstream
func WithRoundTripperExecutor(cfg *Conf) Option {

	nd := &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
		DualStack: true,
	}

	// default transport
	tr := &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		DialContext:           nil,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	return func(dd *defaultDialer) error {

		// Rebuild based on configuration
		tr.MaxIdleConns = fnc(tr.MaxIdleConns, cfg.Tr.MaxIdleConns)
		tr.MaxIdleConnsPerHost = fnc(tr.MaxIdleConnsPerHost, cfg.Tr.MaxIdleConnsPerHost)
		tr.MaxConnsPerHost = fnc(tr.MaxConnsPerHost, cfg.Tr.MaxIdleConnsPerHost)
		tr.IdleConnTimeout = fnt(tr.IdleConnTimeout, cfg.Tr.IdleConnTimeout)
		nd.Timeout = fnt(nd.Timeout, cfg.Nw.Timeout)
		nd.KeepAlive = fnt(nd.KeepAlive, cfg.Nw.KeepAlive)

		// final dialer config
		tr.DialContext = nd.DialContext

		dd.exec = func(
			_ context.Context,
			req *http.Request,
		) (res *http.Response, err error) {
			res, err = tr.RoundTrip(req)
			return
		}

		return nil
	}
}

// WithDefaultExecutor uses default http.DefaultTransport for round trip
func WithDefaultExecutor() Option {
	return func(dd *defaultDialer) error {
		tr := http.DefaultTransport
		dd.exec = func(
			_ context.Context,
			req *http.Request,
		) (*http.Response, error) {
			return tr.RoundTrip(req)
		}
		return nil
	}
}

// WithDefaultValidators sets the validators for the dialer
func WithDefaultValidators() Option {

	return func(dd *defaultDialer) (err error) {
		dd.vals = []validator{
			simpleErrorValidator,
			nilResponseValidator,
			statusValidator,
		}

		return
	}
}

// WithCustomValidator sets custom validator to list of existing validator
func WithCustomValidator(
	fn func(
		context.Context, *http.Request, *http.Response, error,
	) error,
) Option {
	return func(dd *defaultDialer) error {
		dd.vals = append(dd.vals, validator(fn))
		return nil
	}
}

// WithRequestOption sets a custom request option for request
func WithRequestOption(fn RequestOption) Option {
	return func(dd *defaultDialer) (err error) {
		dd.reqopts = append(dd.reqopts, fn)
		return err
	}
}

// WithResponseOption sets a custom response option for request
func WithResponseOption(fn ResponseOption) Option {
	return func(dd *defaultDialer) (err error) {
		dd.resopts = append(dd.resopts, fn)
		return err
	}
}

// WithTimeoutExecutor sets a custom executor which has
// very short timeout
func WithTimeoutExecutor(cfg *TimeoutConf) Option {

	var (
		ex executor
		fn executor
	)

	return func(dd *defaultDialer) (err error) {
		if dd.exec == nil {
			return errors.Wrap(
				errNeedExec, "[dialer.opts] timed",
			)
		}

		ttl := time.Duration(200) * time.Millisecond

		if cfg.Tm > 0 {
			ttl = time.Duration(cfg.Tm) * time.Millisecond
		}

		ex = dd.exec
		fn = func(
			cx context.Context,
			req *http.Request,
		) (res *http.Response, err error) {
			var cf context.CancelFunc

			if ttl > 0 {
				c := req.Context()

				c, cf = context.WithTimeout(
					c, ttl,
				)

				req = req.WithContext(c)
			}

			defer func() {
				if cf != nil {
					cf()
				}
			}()

			res, err = ex(cx, req)
			return
		}

		dd.exec = fn
		return
	}
}
