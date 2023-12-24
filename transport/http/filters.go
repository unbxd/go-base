package http

import (
	"net/http"
)

type Filter func(http.Handler) http.Handler

func Chain(inner http.Handler, filters ...Filter) http.Handler {
	l := len(filters)
	if l == 0 {
		return inner
	}
	return filters[0](Chain(inner, filters[1:]...))
}
