package dialer

import (
	"context"
	"fmt"
	"net/http"
	"time"

	khttp "github.com/go-kit/kit/transport/http"
	"github.com/unbxd/go-base/errors"
)

func fnc(o, n int) int {
	if n > 0 {
		return n
	}
	return o
}

func fnt(o time.Duration, n int) time.Duration {
	if n > 0 {
		return time.Duration(n) * time.Second
	}

	return o
}

func statusCodeError(
	err error,
	prefix string,
	code int,
	requestID interface{},
) error {
	return errors.Wrap(err, fmt.Sprintf(
		"[%s] downstream request failed:[%d] - [%s]: [%v]",
		prefix,
		code,
		http.StatusText(code),
		requestID,
	))

}

// validators
// checks the status code
func statusValidator(
	cx context.Context,
	req *http.Request,
	res *http.Response,
	err error,
) error {
	switch {
	case res.StatusCode >= 500:
		return statusCodeError(
			ErrInternalServer,
			"Downstream has Internal Server Error [net_dialer.v.status_code]",
			res.StatusCode,
			cx.Value(khttp.ContextKeyRequestXRequestID),
		)
	case res.StatusCode == http.StatusNotFound:
		return statusCodeError(
			ErrNotFound,
			"Downstream Says Not Found [net_dialer.v.status_code]",
			res.StatusCode,
			cx.Value(khttp.ContextKeyRequestXRequestID),
		)
	default:
		return nil
	}
}

// checks if there is a simple error
func simpleErrorValidator(
	cx context.Context,
	_ *http.Request,
	_ *http.Response,
	err error,
) error {
	if err != nil {
		return errors.Wrap(ErrExec, err.Error())
	}

	return nil
}

func nilResponseValidator(
	cx context.Context,
	_ *http.Request,
	res *http.Response,
	_ error,
) error {
	if res == nil {
		return ErrResponseIsNil
	}

	return nil
}
