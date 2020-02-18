package http

import "github.com/pkg/errors"

// Standard HTTP Errors
var (
	ErrNotHTTPRequest  = errors.New("missmatch type, request should be *http.Request")
	ErrNotHTTPResponse = errors.New("mismatch type, response should be *http.Response")
)
