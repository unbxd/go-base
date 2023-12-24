package http

import "net/http"

// Default Multiplexer Options

// MethodNotAllowedChiMuxOption sets a custom http.HandlerFunc to handle
// '405' errors in URL Paths
func MethodNotAllowedChiMuxOption(handlerFunc http.HandlerFunc) ChiMuxOption {
	return func(cm *chiMuxer) {
		cm.MethodNotAllowed(handlerFunc)
	}
}

// NotFoundHandlerChiMuxOption sets a custom http.HandlerFunc to handle
// '404' errors when the URL is incorrect
func NotFoundHandlerChiMuxOption(handlerFunc http.HandlerFunc) ChiMuxOption {
	return func(cm *chiMuxer) {
		cm.NotFound(handlerFunc)
	}
}
