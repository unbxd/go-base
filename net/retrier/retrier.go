package retrier

import (
	"context"
	"math/rand"
	net_http "net/http"
	"time"

	"github.com/unbxd/go-base/v2/endpoint"
	"github.com/unbxd/go-base/v2/log"

	"github.com/unbxd/go-base/v2/errors"
	"github.com/unbxd/hystrix-go/hystrix"
)

// States
const (
	PASS State = iota
	FAIL
	RETRY
)

var (
	ErrInternalServer = errors.New("internal server error, response code > 500")
	ErrNotFound       = errors.New("resource not found, response code = 404")
	ErrResponseIsNil  = errors.New("'response' from downstream is nil")
	ErrExec           = errors.New("executor failed")

	ErrRequestIsNotHTTP  = errors.New("retrier request is not net_http.Request")
	ErrResponseIsNotHTTP = errors.New("retrier response is not net_http.Response")
)

type (
	BackoffConf struct {
		Name string
		Incr int
	}

	RetrierConf struct {
		// jitter & classifier can be taken from config
		// in similar fashion
		Backoff *BackoffConf
		// retry counts
		Count int
		// need to set in config,
		// default value is false for bool
		Enable bool
	}

	Deadliner interface {
		Deadline() (time.Duration, error)
	}

	// State defines the state of the connection
	State int

	// Classifier takes a given error generated
	// by the Proxy and assigns a given state based
	// on the error emitted
	Classifier func(error, interface{}) State

	// Backoff defines the strategy in which the duration
	// is computed for the next retry
	Backoff func(counter int) time.Duration

	// Jitter defines the randomization strategy
	// which is added to backoff timer to not clog the
	// the downstream with simulteneous requests
	Jitter func() time.Duration

	Executor func(context.Context, *net_http.Request) (*net_http.Response, error)

	// Retrier retries to perform a single operation
	// multiple times based on the parameters provided
	Retrier struct {
		logger log.Logger

		count  int
		enable bool

		backoff Backoff
		jitter  Jitter
		classfr Classifier

		fn endpoint.Endpoint
	}

	// RetrierOption sets options for Retry
	RetrierOption func(*Retrier) error
)

func (r *Retrier) duration(ctr int) time.Duration {
	return r.backoff(ctr) + r.jitter()
}

func (r *Retrier) Executor() Executor {
	var fn = r.Endpoint()

	return func(
		cx context.Context,
		req *net_http.Request,
	) (res *net_http.Response, err error) {
		rsi, err := fn(cx, req)

		rs, ok := rsi.(*net_http.Response)
		if !ok {
			return nil, ErrResponseIsNotHTTP
		}

		return rs, err
	}
}

// Endpoint returns endpoint.Endpoint with retry wrapped
func (r *Retrier) Endpoint() endpoint.Endpoint {
	return func(
		cx context.Context,
		rqi interface{},
	) (rsi interface{}, err error) {
		if !r.enable {
			return r.fn(cx, rqi)
		}

		var (
			req       Deadliner
			canc      context.CancelFunc
			stamp     = time.Now()
			tolerance = tolerance()()
			ddl       time.Duration
		)

		req = rqi.(Deadliner)

		if ddl, err = req.Deadline(); err == nil {
			// this here is for randomization
			// the request is dropped at the deadline by the
			// Proxy, but the retrier will try again
			//
			// This tolerance will let retrier to try the request
			// again, till it either hits the count limit or
			// reach the deadline computed by the arithmetic
			// request_deadline * tolerance_factor + tolerance_factor

			//TODO check with ujjwal, why multiplication upto 10 on deadline?
			cx, canc = context.WithTimeout(
				cx, time.Duration(
					ddl.Seconds()*
						tolerance+tolerance,
				)*time.Second,
			)
			defer func() {
				canc()
			}()
		}

		r.logger.Debug("Setting UP Retry Loop", log.Int("retry_count", r.count))

		for i := 0; i < r.count; i++ {
			r.logger.Debug(
				"Retrying the endpoint again",
				log.Int("count", i),
				log.Reflect("prev_error", err),
			)

			rsi, err = r.fn(cx, rqi)

			switch cs := r.classfr(err, rsi); cs {
			case PASS, FAIL:
				r.logger.Debug("error classified as PASS/FAIL")

				if err != nil {
					r.logger.Debug(
						"classified as PASS/FAIL with Error",
						log.String("error", err.Error()),
					)
				}

				return rsi, err
			case RETRY:
				r.logger.Debug("error classified as RETRY", log.Reflect("error", err))

				wait := r.duration(i)
				tc := time.After(wait)

				select {
				case <-tc:
					r.logger.Debug(
						"encountered error, retrying",
						log.String("prev-err", err.Error()),
						log.Int64("after", wait.Milliseconds()),
					)
					break
				case <-cx.Done():
					r.logger.Debug(
						"retrier context done. cx.Done()",
						log.Int64(
							"Since",
							time.Since(stamp).Milliseconds(),
						),
					)
					return rsi, err
				}

			default:
				r.logger.Error("This state shouldn't occur", log.Int("classified", int(cs)))
				return rsi, err

			}
		}
		return rsi, err
	}
}

// NewRetrier returns a new Retrier
func NewRetrier(logger log.Logger, fn endpoint.Endpoint, options ...RetrierOption) (*Retrier, error) {
	r := &Retrier{
		fn:      fn,
		classfr: classifier(logger),
		jitter:  jitter(),
		logger:  logger,
	}

	for _, o := range options {
		err := o(r)
		if err != nil {
			return nil, err
		}
	}

	if r.backoff == nil {
		r.backoff = backoff()
	}

	if r.count == 0 {
		r.count = 5
	}

	return r, nil
}

// default classifier
func classifier(logger log.Logger) Classifier {
	return func(err error, res interface{}) State {
		switch {
		// Hysterix Errors
		case err == hystrix.ErrCircuitOpen:
			fallthrough
		case err == hystrix.ErrMaxConcurrency:
			fallthrough
		case err == hystrix.ErrTimeout:
			fallthrough
		case errors.Cause(err) == ErrNotFound:
			logger.Debug("FAILING with Classified ERROR",
				log.String("error", err.Error()),
				log.String("error_cause", errors.Cause(err).Error()),
			)
			return FAIL

		// Errors that will trigger retry
		case errors.Cause(err) == ErrInternalServer:
			fallthrough
		case errors.Cause(err) == ErrResponseIsNil:
			fallthrough
		case errors.Cause(err) == ErrExec:
			logger.Debug("RETRYING with Classified ERROR",
				log.String("error", err.Error()),
				log.String("error_cause", errors.Cause(err).Error()),
			)
			return RETRY

		// No Error, it Pass
		case err == nil:
			logger.Debug("PASSING with No Error, err is nil")
			return PASS
		default:
			logger.Debug(
				"FAILING with unidentified error",
				log.String("error", err.Error()),
				log.String("error_cause", errors.Cause(err).Error()),
			)
			return FAIL
		}
	}
}

// default jitter
func jitter() Jitter {
	rn := rand.New(
		rand.NewSource(
			time.Now().UnixNano(),
		),
	)

	return func() time.Duration {
		return time.Duration(rn.Intn(1000)) * time.Microsecond
	}
}

// default backoff - constant
func backoff() Backoff {
	var incr = 100

	return func(ctr int) time.Duration {
		if ctr <= 0 {
			return 0 * time.Millisecond
		}
		return time.Duration(int64(incr)) * time.Millisecond
	}
}

func tolerance() func() float64 {
	rn := rand.New(
		rand.NewSource(
			time.Now().UnixNano(),
		),
	)

	return func() float64 {
		return float64(rn.Intn(9))
	}
}

// WithLinearBackoff sets backoff as linear
func WithLinearBackoff(conf *BackoffConf) RetrierOption {

	return func(r *Retrier) error {
		var (
			incr int
		)

		if conf.Incr == 0 {
			incr = 100
		} else {
			incr = conf.Incr
		}

		r.backoff = func(ctr int) time.Duration {
			if ctr <= 0 {
				return 0 * time.Millisecond
			}
			return time.Duration(int64(ctr*incr)) * time.Millisecond
		}

		return nil
	}
}

// WithConstantBackoff increments the timer with a constant value
func WithConstantBackoff(conf *BackoffConf) RetrierOption {

	return func(r *Retrier) error {
		var (
			incr int
		)

		if conf.Incr == 0 {
			incr = 100
		} else {
			incr = conf.Incr
		}

		r.backoff = func(ctr int) time.Duration {
			if ctr <= 0 {
				return 0 * time.Millisecond
			}
			return time.Duration(int64(incr)) * time.Millisecond
		}

		return nil
	}
}

// WithRetryCount sets custom retry count for Retrier
func WithRetryCount(count int) RetrierOption {
	return func(r *Retrier) (err error) {
		r.count = count
		return
	}
}

// WithLogger sets the logger for retrier
func WithLogger(logger log.Logger) RetrierOption {
	return func(r *Retrier) (err error) {
		r.logger = logger
		return
	}
}

// WithRetrierEnable sets if Retrier is enabled
func WithRetrierEnable(en bool) RetrierOption {
	return func(r *Retrier) (err error) {
		r.enable = en
		return
	}
}

func WithClassifier(cl Classifier) RetrierOption {
	return func(r *Retrier) (err error) {
		r.classfr = cl
		return
	}
}

// NewRetrierFromConfig returns a new retrier based on configuration
func NewRetrierFromConfig(fn endpoint.Endpoint, lg log.Logger, cfg *RetrierConf, opts ...RetrierOption) (*Retrier, error) {

	opts = append(opts, WithLogger(lg))
	opts = append(opts, WithRetrierEnable(cfg.Enable))

	if cfg.Count > 0 {
		opts = append(opts, WithRetryCount(cfg.Count))
	}
	if cfg.Backoff != nil {
		switch cfg.Backoff.Name {
		case "linear":
			opts = append(opts, WithLinearBackoff(cfg.Backoff))
		case "constant":
			fallthrough
		default:
			opts = append(opts, WithConstantBackoff(cfg.Backoff))
		}
	}

	return NewRetrier(lg, fn, opts...)
}

func toEndpoint(ex Executor) endpoint.Endpoint {
	return func(
		cx context.Context,
		req interface{},
	) (res interface{}, err error) {
		rq, ok := req.(*net_http.Request)
		if !ok {
			return nil, ErrRequestIsNotHTTP
		}

		return ex(cx, rq)
	}
}

func NewExecutorRetrier(
	ex Executor,
	logger log.Logger,
	options ...RetrierOption,
) (*Retrier, error) {
	return NewRetrier(
		logger,
		toEndpoint(ex),
		options...,
	)
}
