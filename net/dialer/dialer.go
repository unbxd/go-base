package dialer

import (
	"context"
	"net/http"
	net_http "net/http"

	"github.com/unbxd/go-base/utils/log"

	"github.com/pkg/errors"
)

// Validator Errors
var (
	ErrInternalServer = errors.New("internal server error, response code > 500")
	ErrNotFound       = errors.New("resource not found, response code = 404")
)

/*
	Dialer and Required Interfaces
*/
type (
	// Dialer interface defines the dialer for a downstream request
	Dialer interface {
		Dial(
			context.Context,
			*net_http.Request,
		) (*net_http.Response, error)
	}
)

type (
	// Option sets optional parameters for dailer
	Option func(*defaultDialer) error

	// RequestOption ...
	RequestOption func(context.Context, *http.Request)

	// ResponseOption ...
	ResponseOption func(context.Context, *http.Response)
)

// NewDialer ...
func NewDialer(logger log.Logger, opts ...Option) (Dialer, error) {
	dd := &defaultDialer{
		lgr:     logger,
		exec:    nil,
		reqopts: []RequestOption{},
		resopts: []ResponseOption{},
		vals:    []validator{},
	}

	for _, o := range opts {
		err := o(dd)
		if err != nil {
			return nil, err
		}
	}

	if dd.exec == nil {
		return nil, errors.Wrap(
			errNeedExec,
			"executory cannot be empty, possible missing options 'WithDefaultExecutor'",
		)
	}

	return dd, nil
}

// NewDefaultDialer returns new default http dialer
func NewDefaultDialer(logger log.Logger, conf *Conf) (Dialer, error) {
	return NewDialer(
		logger,
		WithRoundTripperExecutor(conf),
		WithDefaultValidators(),
	)
}

// NewTimedDialer returns the dialer which times out
func NewTimedDialer(logger log.Logger, conf *Conf) (Dialer, error) {
	return NewDialer(
		logger,
		WithRoundTripperExecutor(conf),
		WithTimeoutExecutor(&conf.To),
	)
}

// dialerError returns an error with context right after sending it to APM
func dialerError(cx context.Context, err error, msg string) error {
	return errors.Wrap(err, msg)
}
