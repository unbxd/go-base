package http

import (
	"github.com/unbxd/go-base/utils/log"
	"go.elastic.co/apm/module/apmhttp"
	net_http "net/http"
)

type Filter func(net_http.Handler) net_http.Handler

func Chain(inner net_http.Handler, filters ...Filter) net_http.Handler {
	l := len(filters)
	if l == 0 {
		return inner
	}
	return filters[0](Chain(inner, filters[1:]...))
}

//very basic panic recovery filter
func PanicRecovery(logger log.Logger) Filter {
	return func(next net_http.Handler) net_http.Handler {
		return net_http.HandlerFunc(func(w net_http.ResponseWriter, r *net_http.Request) {

			defer func() {
				err := recover()
				if err != nil {

					logger.Errorf("panic recovered ", err)

					w.Header().Set("Content-Type", "text/plain")
					w.WriteHeader(net_http.StatusInternalServerError)
					_, err = w.Write([]byte("panic recovered"))
					if err != nil {
						panic(err)
					}
				}

			}()

			next.ServeHTTP(w, r)
		})
	}
}

//elastic apm filter wrapper
func ElasticApm() Filter {
	return func(next net_http.Handler) net_http.Handler {
		return apmhttp.Wrap(next)
	}
}
