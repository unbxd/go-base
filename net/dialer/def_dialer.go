package dialer

import (
	"context"
	"net/http"

	"github.com/unbxd/go-base/utils/log"

	"github.com/pkg/errors"
)

// Error response
var (
	ErrResponseIsNil = errors.New("'response' from downstream is nil")
	ErrExec          = errors.New("executor failed")

	errNeedExec = errors.New("needs existing executor")
)

type (
	// Executes the downstream call
	executor func(context.Context, *http.Request) (*http.Response, error)
	// validates the downstream response returned
	validator func(context.Context, *http.Request, *http.Response, error) error
)

type (
	defaultDialer struct {
		lgr  log.Logger
		exec executor

		reqopts []RequestOption
		resopts []ResponseOption

		vals []validator
	}
)

// Dial methods wraps options and dials downstream
func (dd *defaultDialer) Dial(
	cx context.Context,
	req *http.Request,
) (res *http.Response, err error) {
	// request decorator
	for _, fn := range dd.reqopts {
		fn(cx, req)
	}

	if dd.exec == nil {
		return nil, errors.Wrap(
			errNeedExec,
			"executory cannot be empty, possible missing options 'WithDefaultExecutor'",
		)
	}

	// execute the downstream
	res, err = dd.exec(cx, req)

	// validate the respons/err
	for _, fn := range dd.vals {
		er := fn(cx, req, res, err)
		if er != nil {
			return res, dialerError(
				cx, er, "validation failed",
			)
		}
	}

	// if all looks good, decorate response
	for _, fn := range dd.resopts {
		fn(cx, res)
	}

	return
}
