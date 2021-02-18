package http

import (
	"fmt"
	"github.com/unbxd/go-base/utils/log"
	"go.elastic.co/apm/module/apmhttp"
	net_http "net/http"
	"runtime"
	"runtime/debug"
	"text/template"
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

/** For DecoratedPanicFormatter, Ported from negroni/recovery **/
/** https://github.com/urfave/negroni/blob/master/recovery.go **/

const (
	// NoPrintStackBodyString is the body content returned when HTTP stack printing is suppressed
	NoPrintStackBodyString = "500 Internal Server Error"

	panicText = "PANIC: %s\n%s"
	panicHTML = `<html>
<head><title>PANIC: {{.RecoveredPanic}}</title></head>
<style type="text/css">
html, body {
	font-family: Helvetica, Arial, Sans;
	color: #333333;
	background-color: #ffffff;
	margin: 0px;
}
h1 {
	color: #ffffff;
	background-color: #f14c4c;
	padding: 20px;
	border-bottom: 1px solid #2b3848;
}
.block {
	margin: 2em;
}
.panic-interface {
}
.panic-stack-raw pre {
	padding: 1em;
	background: #f6f8fa;
	border: dashed 1px;
}
.panic-interface-title {
	font-weight: bold;
}
</style>
<body>
<h1>Negroni - PANIC</h1>
<div class="panic-interface block">
	<h3>{{.RequestDescription}}</h3>
	<span class="panic-interface-title">Runtime error:</span> <span class="panic-interface-element">{{.RecoveredPanic}}</span>
</div>
{{ if .Stack }}
<div class="panic-stack-raw block">
	<h3>Runtime Stack</h3>
	<pre>{{.StackAsString}}</pre>
</div>
{{ end }}
</body>
</html>`
)

type (

	// PanicInformation contains all
	// elements for printing stack informations.
	PanicInformation struct {
		RecoveredPanic interface{}
		Stack          []byte
		Request        *net_http.Request
	}

	// PanicFormatter is an interface on object can implement
	// to be able to output the stack trace
	PanicFormatter interface {
		// FormatPanicError output the stack for a given answer/response.
		// In case the the middleware should not output the stack trace,
		// the field `Stack` of the passed `PanicInformation` instance equals `[]byte{}`.
		FormatPanicError(rw net_http.ResponseWriter, r *net_http.Request, infos *PanicInformation)
	}

	TextPanicFormatter struct{ tt string }
	HTMLPanicFormatter struct{ ttl *template.Template }

	// Recovery is a Negroni middleware that recovers from any panics and writes a 500 if there was one.
	Recovery struct {
		Logger           log.Logger
		PrintStack       bool
		LogStack         bool
		PanicHandlerFunc func(*PanicInformation)
		StackAll         bool
		StackSize        int
		Formatter        PanicFormatter
	}

	RecoveryFilterOption func(*Recovery)
)

// StackAsString returns a printable version of the stack
func (p *PanicInformation) StackAsString() string {
	return string(p.Stack)
}

// RequestDescription returns a printable description of the url
func (p *PanicInformation) RequestDescription() string {

	if p.Request == nil {
		return "request is nil"
	}

	var queryOutput string
	if p.Request.URL.RawQuery != "" {
		queryOutput = "?" + p.Request.URL.RawQuery
	}
	return fmt.Sprintf("%s %s%s", p.Request.Method, p.Request.URL.Path, queryOutput)
}

func (t *HTMLPanicFormatter) FormatPanicError(rw net_http.ResponseWriter, r *net_http.Request, infos *PanicInformation) {
	if rw.Header().Get("Content-Type") == "" {
		rw.Header().Set("Content-Type", "text/html; charset=utf-8")
	}
	t.ttl.Execute(rw, infos)
}

func (t *TextPanicFormatter) FormatPanicError(rw net_http.ResponseWriter, r *net_http.Request, infos *PanicInformation) {
	if rw.Header().Get("Content-Type") == "" {
		rw.Header().Set("Content-Type", "text/plain; charset=utf-8")
	}
	fmt.Fprintf(rw, t.tt, infos.RecoveredPanic, infos.Stack)
}

// WithProductionDefaultsRecoveryFilterOption configures the Recovery filter
// for default Production configuration
func WithProductionDefaultsRecoveryFilterOption() RecoveryFilterOption {
	return func(r *Recovery) {
		r.PrintStack = false
		r.LogStack = true
		r.Formatter = &TextPanicFormatter{panicText}
	}
}

// WithDevelopmentDefaultsRecoveryFilterOption configures the Recovery filter
// for default development configuration
func WithDevelopmentDefaultsRecoveryFilterOption() RecoveryFilterOption {
	return func(r *Recovery) {
		r.PrintStack = true
		r.LogStack = true
		r.Formatter = &HTMLPanicFormatter{
			template.Must(template.New("PanicPage").Parse(panicHTML)),
		}
	}
}

// WithStackSizeRecoveryFilterOption configures the stack size returned
func WithStackSizeRecoveryFilterOption(ss int) RecoveryFilterOption {
	return func(r *Recovery) { r.StackSize = ss }
}

// WithCustomFormatterRecoveryFilterOption sets a custom formatter for
// recovery to use
func WithCustomFormatterRecoveryFilterOption(fmt PanicFormatter) RecoveryFilterOption {
	return func(r *Recovery) { r.Formatter = fmt }
}

// WithPrintStackRecoveryFilterOption lets you configure if the stack
// is to be printed or not
func WithPrintStackRecoveryFilterOption(ps bool) RecoveryFilterOption {
	return func(r *Recovery) { r.PrintStack = ps }
}

// WithPanicHandlerFuncRecoveryFilterOption lets the panic handler function
// to be configured
func WithPanicHandlerFuncRecoveryFilterOption(fn func(*PanicInformation)) RecoveryFilterOption {
	return func(r *Recovery) { r.PanicHandlerFunc = fn }
}

// NewRecovery returns a new instance of Recovery
func NewRecovery(logger log.Logger, options ...RecoveryFilterOption) *Recovery {
	r := &Recovery{
		Logger:     logger,
		PrintStack: true,
		LogStack:   true,
		StackAll:   false,
		StackSize:  1024 * 8,
		Formatter:  &TextPanicFormatter{},
	}

	for _, o := range options {
		o(r)
	}

	return r
}
func DecoratedPanicRecoveryFilter(logger log.Logger, opts ...RecoveryFilterOption) Filter {
	rec := NewRecovery(logger, opts...)
	return func(next net_http.Handler) net_http.Handler {
		return net_http.HandlerFunc(func(w net_http.ResponseWriter, r *net_http.Request) {

			defer func() {
				err := recover()
				if err != nil {
					// set status code
					w.WriteHeader(net_http.StatusInternalServerError)

					infos := &PanicInformation{
						RecoveredPanic: err,
						Request:        r,
						Stack:          make([]byte, rec.StackSize),
					}

					infos.Stack = infos.Stack[:runtime.Stack(infos.Stack, rec.StackAll)]

					if rec.PrintStack && rec.Formatter != nil {
						rec.Formatter.FormatPanicError(w, r, infos)
					} else {
						if w.Header().Get("Content-Type") == "" {
							w.Header().Set("Content-Type", "text/plain; charset=utf-8")
						}
						fmt.Fprint(w, NoPrintStackBodyString)
					}

					if rec.LogStack {
						rec.Logger.Errorf(panicText, err, infos.Stack)
					}

					if rec.PanicHandlerFunc != nil {
						func() {
							defer func() {
								if err := recover(); err != nil {
									rec.Logger.Errorf("provided PanicHandlerFunc panic'd: %s, trace:\n%s", err, debug.Stack())
									rec.Logger.Errorf("%s\n", debug.Stack())
								}
							}()
							rec.PanicHandlerFunc(infos)
						}()
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
