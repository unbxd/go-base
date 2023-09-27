package http

import (
	"context"

	net_http "net/http"

	kit_http "github.com/go-kit/kit/transport/http"
)

// NewRequestIDHandlerOption returns a HandlerOption for a customheader to be populated
// with request id, generated at filter
// This is same as CustomRequestIDFilter except at per Handler level
func NewRequestIDHandlerOption(formatter RequestIDFormatter, customHeaders ...string) HandlerOption {
	return func(h *handler) {
		h.filters = append(
			h.filters,
			CustomRequestIDFilter(formatter, customHeaders...),
		)
	}
}

// NewDeleteHeaderHandlerOption deletes the headers from net_http.Request
// before it is sent to HandlerFunc
func NewDeleteHeaderHandlerOption(headers ...string) HandlerOption {
	return func(h *handler) {
		h.options = append(h.options, kit_http.ServerBefore(
			func(ctx context.Context, req *net_http.Request) context.Context {

				for _, h := range headers {
					req.Header.Del(h)
				}

				return ctx
			},
		))
	}
}

// NewSetRequestHeader sets the request header
func NewSetRequestHeader(key, val string) HandlerOption {
	return func(h *handler) {
		h.options = append(
			h.options,
			kit_http.ServerBefore(kit_http.SetRequestHeader(key, val)),
		)
	}
}

// NewSetResponseHeader sets the response header
func NewSetResponseHeader(key, val string) HandlerOption {
	return func(h *handler) {
		h.options = append(
			h.options,
			kit_http.ServerAfter(kit_http.SetResponseHeader(key, val)),
		)
	}
}

// NewFiltersHandlerOption allows custom filter added per route
func NewFiltersHandlerOption(filters ...Filter) HandlerOption {
	return func(h *handler) {
		h.filters = append(h.filters, filters...)
	}
}
