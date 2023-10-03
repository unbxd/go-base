package http

import (
	"bytes"
	"io"
	net_http "net/http"
)

// ResponseOption defines the options which modify the
// net/http Response
type ResponseOption func(*net_http.Response)

// ResponseWithBytes provide option to update the response body
// with Bytes data
func ResponseWithBytes(bt []byte) ResponseOption {
	return func(res *net_http.Response) {
		res.Body = io.NopCloser(bytes.NewReader(bt))
	}
}

// ResponseWithCode provides option to update the status code for the
// reponse
func ResponseWithCode(code int) ResponseOption {
	return func(res *net_http.Response) {
		res.Status = net_http.StatusText(code)
		res.StatusCode = code
	}
}

// ResponseWithReader provies option to update the body
// of the response with a custom reader
func ResponseWithReader(reader io.Reader) ResponseOption {
	return func(res *net_http.Response) {
		res.Body = io.NopCloser(reader)
	}
}

// NewResponse returns a new net/http.Response based on incoming
// request and the available options passed to it
func NewResponse(req *net_http.Request, opts ...ResponseOption) *net_http.Response {
	r := &net_http.Response{
		Status:     "undefined",
		StatusCode: 0,
		Request:    req,
		Body:       nil,
	}

	for _, o := range opts {
		o(r)
	}

	return r
}
