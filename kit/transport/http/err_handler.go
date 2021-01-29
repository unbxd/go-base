package http

import "github.com/go-kit/kit/transport"

// ErrorHandler wraps onn top of transport.ErrorHandler and provides
// a hook for diagnostic purpose
type ErrorHandler interface {
	transport.ErrorHandler
}
