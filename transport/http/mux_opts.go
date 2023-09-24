package http

import "net/http"

// Default Multiplexer Options

// MethodNotAllowedMuxOption sets a custom http.HandlerFunc to handle
// '405' errors in URL Paths
func MethodNotAllowedMuxOption(handlerFunc http.HandlerFunc) DefaultMuxOption {
	return func(cm *chiMuxer) {
		cm.MethodNotAllowed(handlerFunc)
	}
}

// NotFoundHandlerMuxOption sets a custom http.HandlerFunc to handle
// '404' errors when the URL is incorrect
func NotFoundHandlerMuxOption(handlerFunc http.HandlerFunc) DefaultMuxOption {
	return func(cm *chiMuxer) {
		cm.NotFound(handlerFunc)
	}
}
