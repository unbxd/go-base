package http

import (
	"fmt"
	"html/template"
	"net/http"
	"runtime"

	"github.com/unbxd/go-base/v2/log"
)

const (
	TextPanicFormatter PanicFormatterType = iota
	HTMLPanicFormatter
)

type (
	PanicFormatterType uint
	// PanicInformation holds all elements required to print
	// stack information about the panic
	PanicInformation struct {
		RecoveredPanic interface{}
		Stack          []byte
		Request        *http.Request
	}

	// PanicFormatter provides an interface to print stack trace
	// in customized way
	PanicFormatter interface {
		Format(http.ResponseWriter, *http.Request, *PanicInformation)
	}

	// types of panic formatters
	textPanicFormatter struct{}
	htmlPanicFormatter struct{ template *template.Template }
	// TODO: JSON Formatter

	// PanicCallback gives a callback option to handle Panic with details
	PanicCallback func(*PanicInformation)

	recovery struct {
		returnStack bool

		stackSize   int
		stackOthers bool // Stack formats stack traces of all other goroutines into buf after the trace for the current goroutine

		logger    log.Logger
		formatter PanicFormatter
	}

	RecoveryOption func(*recovery)
)

// PanicInformation
func (p *PanicInformation) StackString() string {
	return string(p.Stack)
}

func (p *PanicInformation) RequestInfo() string {

	if p.Request == nil {
		return "request: is nil"
	}

	var queryOutput string
	if p.Request.URL.RawQuery != "" {
		queryOutput = "?" + p.Request.URL.RawQuery
	}
	return fmt.Sprintf("%s %s%s", p.Request.Method, p.Request.URL.Path, queryOutput)
}

// Formatter Methods
// textFormatter
func (text *textPanicFormatter) Format(w http.ResponseWriter, r *http.Request, info *PanicInformation) {
	// force a content-type
	w.Header().Set(HeaderContentType, "text/plain; charset=utf-8")
	fmt.Fprintf(w, "PANIC: %s\n%s", info.RecoveredPanic, info.Stack)
}

func newTextPanicFormatter() PanicFormatter { return &textPanicFormatter{} }

// htmlFormatter
func (html *htmlPanicFormatter) Format(w http.ResponseWriter, r *http.Request, info *PanicInformation) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = html.template.Execute(w, info)
}

func newHtmlPanicFormatter() PanicFormatter {
	txt := `
	<html><head><title>PANIC: {{.RecoveredPanic}}</title></head>
	<style type="text/css">body,html{font-family:Helvetica,Arial,Sans;color:#333;background-color:#fff;margin:0}h1{color:#fff;background-color:#f14c4c;padding:20px;border-bottom:1px solid #2b3848}.block{margin:2em}.panic-stack-raw pre{padding:1em;background:#f6f8fa;border:1px dashed}.panic-interface-title{font-weight:700}</style>
	<body><h1>Negroni - PANIC</h1>
	<div class="panic-interface block">
	<h3>{{.RequestDescription}}</h3>
	<span class="panic-interface-title">Runtime error:</span> <span class="panic-interface-element">{{.RecoveredPanic}}</span>
	</div>
	{{ if .Stack }}
	<div class="panic-stack-raw block"><h3>Runtime Stack</h3><pre>{{.StackAsString}}</pre></div>
	{{ end }}
	</body></html>
	`
	tt := template.Must(template.New("PanicPage").Parse(txt))
	return &htmlPanicFormatter{tt}
}

func (rec *recovery) HandlerFunc(
	next http.Handler,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.WriteHeader(http.StatusInternalServerError)

				info := &PanicInformation{
					RecoveredPanic: err,
					Request:        r,
					Stack:          make([]byte, rec.stackSize),
				}

				defer func() {
					if rec.logger == nil {
						fmt.Printf(
							"Panic: Internal Server Error \nerror: %v\nstack: %v",
							err, info.StackString(),
						)
						return
					}

					// we always log stack trace
					rec.logger.Error(
						"panic: internal server error",
						log.Reflect("error", err),
						log.String("stackTrace", info.StackString()),
					)
				}()

				if rec.returnStack {
					info.Stack = info.Stack[:runtime.Stack(info.Stack, rec.stackOthers)]
				}

				// if we don't have formatter, but we need stack
				if rec.formatter == nil {
					newTextPanicFormatter().Format(w, r, info)
					return
				}

				rec.formatter.Format(w, r, info)
				return
			}
		}()
		next.ServeHTTP(w, r)
	}
}

func WithTextFormatter() RecoveryOption {
	return func(r *recovery) { r.formatter = newTextPanicFormatter() }
}

func WithHTMLFormatter() RecoveryOption {
	return func(r *recovery) { r.formatter = newHtmlPanicFormatter() }
}

func WithCustomFormatter(formatter PanicFormatter) RecoveryOption {
	return func(r *recovery) { r.formatter = formatter }
}

func WithStack(stackSize int, stackOtherGoroutines bool) RecoveryOption {
	return func(r *recovery) {
		r.returnStack = true
		r.stackSize = stackSize
		r.stackOthers = false
	}
}

func WithFormatterType(formatterType PanicFormatterType) RecoveryOption {
	return func(r *recovery) {
		switch formatterType {
		case TextPanicFormatter:
			WithTextFormatter()(r)
		case HTMLPanicFormatter:
			WithHTMLFormatter()(r)
		}
	}
}

func WithoutStack() RecoveryOption {
	return func(r *recovery) { r.returnStack = false; r.stackOthers = false }
}

func NewRecovery(
	logger log.Logger,
	options ...RecoveryOption,
) *recovery {
	r := &recovery{
		returnStack: false,
		stackSize:   1024 * 8,
		stackOthers: false,
		logger:      logger,
		formatter:   nil,
	}

	for _, o := range options {
		o(r)
	}

	if r.formatter == nil {
		r.formatter = &textPanicFormatter{}
	}

	return r
}

func PanicRecoveryFilter(logger log.Logger, options ...RecoveryOption) Filter {
	recovery := NewRecovery(logger, options...)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(recovery.HandlerFunc(next))
	}
}
