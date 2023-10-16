package notifier

import (
	"context"
	"sync"
	"time"

	"github.com/unbxd/go-base/v2/errors"
	"github.com/unbxd/go-base/v2/log"
	"github.com/unbxd/go-base/v2/transport/nats"
)

type (
	writer interface {
		Write(cx context.Context, data interface{}) error
	}

	defaultWriter struct{ *natsNotifier }

	bufferedWriter struct {
		*natsNotifier

		logger log.Logger

		buffer []interface{}

		mu sync.Mutex
	}
)

// Default Writer
func (dw *defaultWriter) Write(cx context.Context, data interface{}) error {
	return dw.Publish(cx, dw.natsNotifier.subject, data)
}

func newDefaultWriter(nn *natsNotifier) writer { return &defaultWriter{nn} }

// Write writes the data to the buffer
func (bw *bufferedWriter) Write(cx context.Context, data interface{}) (err error) {
	bw.mu.Lock()
	bw.buffer = append(bw.buffer, data)
	bw.mu.Unlock()
	return
}

// Producer runs periodically and dumps all the events on the returned channel
func (bw *bufferedWriter) Producer(
	periodicity time.Duration,
	done <-chan struct{},
) (<-chan interface{}, chan error) {
	// create channel for generating data
	var (
		datas  = make(chan interface{})
		errch  = make(chan error)
		ticker = time.NewTicker(periodicity)
	)

	go func() {
		for {
			select {
			case <-ticker.C:
				// perform operation of reading the buffer
				// and writing the data from buffer in channel
				bw.mu.Lock()
				newBuffer := make([]interface{}, len(bw.buffer))
				copy(newBuffer, bw.buffer)
				bw.buffer = nil
				bw.mu.Unlock() // release lock as soon as possible

				// iterate over newBuffer and emit on channel
				for _, data := range newBuffer {
					datas <- data
				}
			case err := <-errch:
				// error recieved from worker
				bw.logger.Error(
					"Got error from Worker:",
					log.String("error", err.Error()),
					log.Error(err),
				)
			case <-done:
				close(datas)
				close(errch)
				return
			}
		}
	}()
	return datas, errch
}

// Buffered Writer
func (bw *bufferedWriter) Worker(
	done <-chan struct{},
	datas <-chan interface{},
	errch chan<- error,
) {
	go func() {
		for {
			select {
			case data := <-datas:
				cx := context.Background()
				err := bw.Publish(cx, bw.subject, data)
				if err != nil {
					errch <- err
				}
			case <-done:
				return
			}

		}
	}()
}

func newBufferedWriter(
	logger log.Logger,
	bufferSize int,
	parallelism int,
	periodicity time.Duration,
	nn *natsNotifier,
) writer {

	done := make(chan struct{})
	// start producer
	bw := &bufferedWriter{
		natsNotifier: nn,
		logger:       logger,
		buffer:       make([]interface{}, 0),
	}

	datach, errch := bw.Producer(periodicity, done)

	for i := 0; i < parallelism; i++ {
		go func() {
			bw.Worker(done, datach, errch)
		}()
	}

	return bw
}

type (
	Option  func(*natsNotifier)
	Encoder nats.PublishMessageEncoder

	natsNotifier struct {
		*nats.Publisher

		writer writer

		opts    []nats.PublisherOption
		subject string
	}
)

func (nn *natsNotifier) Notify(
	cx context.Context, data interface{},
) error {
	return nn.writer.Write(
		cx,
		data,
	)
}

// Options

func WithSubjectPrefix(prefix string) Option {
	return func(nn *natsNotifier) {
		nn.opts = append(
			nn.opts,
			nats.WithPublisherSubjectPrefix(prefix),
		)
	}
}

func WithMessageEncoder(fn Encoder) Option {
	return func(p *natsNotifier) {
		p.opts = append(
			p.opts,
			nats.WithPublishMessageEncoder(
				nats.PublishMessageEncoder(fn),
			),
		)
	}
}

func WithBufferedWriter(
	logger log.Logger,
	bufferSize int,
	parallelism int,
	periodicity time.Duration,
) Option {
	return func(nn *natsNotifier) {
		nn.writer = newBufferedWriter(
			logger, bufferSize, parallelism, periodicity, nn,
		)
	}
}

// NewNotifier returns a default implementation of Notifier, which
// relies on NATS to publish the events.
// Any future implementation should name itself as `New<type>Notifier`
func NewNotifier(
	connstr string,
	subject string,
	options ...Option,
) (Notifier, error) {
	var nn *natsNotifier

	nn = &natsNotifier{
		subject: subject,
		writer:  newDefaultWriter(nn),
	}

	for _, fn := range options {
		fn(nn)
	}

	pub, err := nats.NewPublisher(
		connstr, nn.opts...,
	)

	if err != nil {
		return nil, errors.Wrap(
			err,
			"failed to create publisher",
		)
	}

	nn.Publisher = pub
	return nn, nil
}
