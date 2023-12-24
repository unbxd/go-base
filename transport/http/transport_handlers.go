package http

import net_http "net/http"

// Get handles GET request
func (tr *Transport) Get(url string, fn HandlerFunc, options ...HandlerOption) {
	tr.muxer.Handler(net_http.MethodGet, url, encapsulate(fn, tr.handlerOptions, options))
}

// GET provides flexible interface for handling request for GET method
// It exposes a structured logical break up of the function handling
// the request.
// Breakup consists of
// - decoder (decodes the request & converts to business object)
// - endpoint (handles the business object, and returns a result in business object )
// - encoder (converts business object into response)
func (tr *Transport) GET(
	uri string,
	fn Handler,
	options ...HandlerOption,
) {
	tr.muxer.Handler(
		net_http.MethodGet,
		uri,
		newHandler(fn, append(tr.handlerOptions, options...)...),
	)
}

// Put handles PUT request
func (tr *Transport) Put(url string, fn HandlerFunc, options ...HandlerOption) {
	tr.muxer.Handler(net_http.MethodPut, url, encapsulate(fn, tr.handlerOptions, options))
}

// PUT provides flexible interface for handling request for put method
// it exposes a structured logical break up of the function handling
// the request.
// breakup consists of
// - decoder (decodes the request & converts to business object)
// - endpoint (handles the business object, and returns a result in business object )
// - encoder (converts business object into response)
func (tr *Transport) PUT(
	url string,
	fn Handler,
	options ...HandlerOption,
) {
	tr.muxer.Handler(
		net_http.MethodPut,
		url,
		newHandler(fn, append(tr.handlerOptions, options...)...),
	)
}

// Post handles POST request
func (tr *Transport) Post(url string, fn HandlerFunc, options ...HandlerOption) {
	tr.muxer.Handler(net_http.MethodPost, url, encapsulate(fn, tr.handlerOptions, options))
}

// POST provides flexible interface for handling request for post method
// it exposes a structured logical break up of the function handling
// the request.
// breakup consists of
// - decoder (decodes the request & converts to business object)
// - endpoint (handles the business object, and returns a result in business object )
// - encoder (converts business object into response)
func (tr *Transport) POST(
	url string,
	fn Handler,
	options ...HandlerOption,
) {
	tr.muxer.Handler(
		net_http.MethodPost,
		url,
		newHandler(fn, append(tr.handlerOptions, options...)...),
	)
}

// Delete handles DELETE request
func (tr *Transport) Delete(url string, fn HandlerFunc, options ...HandlerOption) {
	tr.muxer.Handler(net_http.MethodDelete, url, encapsulate(fn, tr.handlerOptions, options))
}

// DELETE provides flexible interface for handling request for delete method
// it exposes a structured logical break up of the function handling
// the request.
// breakup consists of
// - decoder (decodes the request & converts to business object)
// - endpoint (handles the business object, and returns a result in business object )
// - encoder (converts business object into response)
func (tr *Transport) DELETE(
	url string,
	fn Handler,
	options ...HandlerOption,
) {
	tr.muxer.Handler(
		net_http.MethodDelete,
		url,
		newHandler(fn, append(tr.handlerOptions, options...)...),
	)
}

// Patch handles PATCH request
func (tr *Transport) Patch(url string, fn HandlerFunc, options ...HandlerOption) {
	tr.muxer.Handler(net_http.MethodPatch, url, encapsulate(fn, tr.handlerOptions, options))
}

// PATCH provides flexible interface for handling request for patch method
// it exposes a structured logical break up of the function handling
// the request.
// breakup consists of
// - decoder (decodes the request & converts to business object)
// - endpoint (handles the business object, and returns a result in business object )
// - encoder (converts business object into response)
func (tr *Transport) PATCH(
	url string,
	fn Handler,
	options ...HandlerOption,
) {
	tr.muxer.Handler(
		net_http.MethodPatch,
		url,
		newHandler(fn, append(tr.handlerOptions, options...)...),
	)
}

// Options handles OPTIONS request
func (tr *Transport) Options(url string, fn HandlerFunc, options ...HandlerOption) {
	tr.muxer.Handler(net_http.MethodOptions, url, encapsulate(fn, tr.handlerOptions, options))
}

// OPTION provides flexible interface for handling request for option method
// it exposes a structured logical break up of the function handling
// the request.
// breakup consists of
// - decoder (decodes the request & converts to business object)
// - endpoint (handles the business object, and returns a result in business object )
// - encoder (converts business object into response)
func (tr *Transport) OPTION(
	url string,
	fn Handler,
	options ...HandlerOption,
) {
	tr.muxer.Handler(
		net_http.MethodOptions,
		url,
		newHandler(fn, append(tr.handlerOptions, options...)...),
	)
}

// Head handles HEAD request
func (tr *Transport) Head(url string, fn HandlerFunc, options ...HandlerOption) {
	tr.muxer.Handler(net_http.MethodHead, url, encapsulate(fn, tr.handlerOptions, options))
}

// HEAD provides flexible interface for handling request for head method
// it exposes a structured logical break up of the function handling
// the request.
// breakup consists of
// - decoder (decodes the request & converts to business object)
// - endpoint (handles the business object, and returns a result in business object )
// - encoder (converts business object into response)
func (tr *Transport) HEAD(
	url string,
	fn Handler,
	options ...HandlerOption,
) {
	tr.muxer.Handler(
		net_http.MethodHead,
		url,
		newHandler(fn, append(tr.handlerOptions, options...)...),
	)
}

// Trace handles TRACE request
func (tr *Transport) Trace(url string, fn HandlerFunc, options ...HandlerOption) {
	tr.muxer.Handler(net_http.MethodTrace, url, encapsulate(fn, tr.handlerOptions, options))
}

// TRACE provides flexible interface for handling request for trace method
// it exposes a structured logical break up of the function handling
// the request.
// breakup consists of
// - decoder (decodes the request & converts to business object)
// - endpoint (handles the business object, and returns a result in business object )
// - encoder (converts business object into response)
func (tr *Transport) TRACE(
	url string,
	fn Handler,
	options ...HandlerOption,
) {
	tr.muxer.Handler(
		net_http.MethodTrace,
		url,
		newHandler(fn, append(tr.handlerOptions, options...)...),
	)
}

// Handle is generic method to allow custom bindings of URL with a method and it's handler
func (tr *Transport) Handle(method, url string, fn HandlerFunc, options ...HandlerOption) {
	tr.muxer.Handler(method, url, encapsulate(fn, tr.handlerOptions, options))
}

// HANDLE gives a generic method agnostic way of binding handler to the request
func (tr *Transport) HANDLE(met, url string, fn Handler, options ...HandlerOption) {
	tr.muxer.Handler(
		met, url,
		newHandler(fn, append(tr.handlerOptions, options...)...),
	)
}
