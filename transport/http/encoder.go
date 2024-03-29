package http

import (
	"context"
	"fmt"
	"io"
	net_http "net/http"
	"sync"
	"time"

	kit_http "github.com/go-kit/kit/transport/http"
	"github.com/oxtoacart/bpool"
)

// Encoder denotes the Encoder used to write the data on stream
// after reading the interface
type Encoder func(context.Context, net_http.ResponseWriter, interface{}) error

// A pool is an interface for getting and returning temporary
// byte slices for use by io.CopyBuffer.
type pool interface {
	Get() []byte
	Put([]byte)
}

type flusher interface {
	io.Writer
	net_http.Flusher
}

type latencyWriter struct {
	dst     flusher
	latency time.Duration
	mu      sync.Mutex
	timer   *time.Timer
	pending bool
}

func (lw *latencyWriter) Write(p []byte) (n int, err error) {
	lw.mu.Lock()
	defer lw.mu.Unlock()

	n, err = lw.dst.Write(p)

	if lw.latency < 0 {
		lw.dst.Flush()
		return
	}

	if lw.pending {
		return
	}

	if lw.timer == nil {
		lw.timer = time.AfterFunc(lw.latency, lw.delayedFlush)
	} else {
		lw.timer.Reset(lw.latency)
	}

	lw.pending = true

	return
}

func (lw *latencyWriter) delayedFlush() {
	lw.mu.Lock()
	defer lw.mu.Unlock()

	if !lw.pending {
		return
	}

	lw.dst.Flush()
	lw.pending = false
}

func (lw *latencyWriter) stop() {
	lw.mu.Lock()
	defer lw.mu.Unlock()

	lw.pending = false
	if lw.timer != nil {
		lw.timer.Stop()
	}
}

// util methods to copy response from *net_http.Response to net_http.ResponseWriter

// copies response from response.Body to ResponseWriter
func copyResponse(
	bp pool,
	dst io.Writer,
	src io.Reader,
	flushdur time.Duration,
) error {
	if flushdur != 0 {
		if wm, ok := dst.(flusher); ok {
			lw := &latencyWriter{
				dst:     wm,
				latency: flushdur,
			}

			defer lw.stop()

			lw.pending = true
			lw.timer = time.AfterFunc(flushdur, lw.delayedFlush)

			dst = lw
		}
	}

	var buf []byte
	if bp != nil {
		buf = bp.Get()
		defer bp.Put(buf)
	}

	_, err := copyBuffer(dst, src, buf)
	return err
}

// copyBuffer returns any write errors or non-EOF read errors, and the amount
// of bytes written.
func copyBuffer(dst io.Writer, src io.Reader, buf []byte) (int64, error) {
	if len(buf) == 0 {
		buf = make([]byte, 32*1024)
	}
	var written int64
	for {
		nr, rerr := src.Read(buf)
		if rerr != nil && rerr != io.EOF && rerr != context.Canceled {
			return written, fmt.Errorf("read error during body copy: %v", rerr)

		}
		if nr > 0 {
			nw, werr := dst.Write(buf[:nr])
			if nw > 0 {
				written += int64(nw)
			}
			if werr != nil {
				return written, werr
			}
			if nr != nw {
				return written, io.ErrShortWrite
			}
		}
		if rerr != nil {
			if rerr == io.EOF {
				rerr = nil
			}
			return written, rerr
		}
	}
}

func flushInterval(res *net_http.Response) time.Duration {
	resCT := res.Header.Get("Content-Type")

	// For Server-Sent Events responses, flush immediately.
	// The MIME type is defined in https://www.w3.org/TR/eventsource/#text-event-stream
	if resCT == "text/event-stream" {
		return -1 // negative means immediately
	}

	// TODO: more specific cases?
	return 10 * time.Millisecond
}

func copyHeader(dst, src net_http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

func newDefaultEncoder() Encoder {
	bufferPool := bpool.NewBytePool(100, 1000000)
	return func(ctx context.Context, rw net_http.ResponseWriter, res interface{}) (err error) {
		rr, ok := res.(*net_http.Response)
		if !ok {
			return ErrNotHTTPResponse
		}

		if res == nil {
			rw.WriteHeader(net_http.StatusNoContent)
			return
		}

		copyHeader(rw.Header(), rr.Header)

		switch {
		case rr.StatusCode == 0:
			rw.WriteHeader(net_http.StatusOK)
		case rr.StatusCode > 0:
			rw.WriteHeader(rr.StatusCode)
		default:
			panic("status code should be non-negative")
		}

		defer func() {
			rr.Body.Close()
			rr.Close = true
		}()

		return copyResponse(bufferPool, rw, rr.Body, flushInterval(rr))
	}
}

// NewDefaultEncoder returns a default Encoder used by http
func NewDefaultEncoder() Encoder { return newDefaultEncoder() }

// NewDefaultJSONEncoder encodes the response in JSON
func NewDefaultJSONEncoder() Encoder { return Encoder(kit_http.EncodeJSONResponse) }

// NewGoKitEncoderHandlerOption provides option to encode the request
func NewGoKitEncoderHandlerOption(fn kit_http.EncodeResponseFunc) HandlerOption {
	return func(h *handler) {
		h.encoder = Encoder(fn)
	}
}
