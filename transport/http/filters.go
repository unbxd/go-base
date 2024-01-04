package http

import (
	"net/http"
)

type Filter func(http.Handler) http.Handler

func chain(inner http.Handler, filters ...Filter) http.Handler {
	l := len(filters)
	if l == 0 {
		return inner
	}
	return filters[0](chain(inner, filters[1:]...))
}
