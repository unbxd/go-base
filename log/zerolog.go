package log

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/diode"
)

type (
	zeroLogger struct {
		withStack bool
		logger    zerolog.Logger
	}

	zeroLoggerConfig struct {
		level      zerolog.Level
		writer     io.Writer
		withCaller bool
		withStack  bool

		fields []Field
	}

	ZeroLoggerOption func(*zeroLoggerConfig) error
)

func zerologEventFields(event *zerolog.Event, fields ...Field) *zerolog.Event {
	for _, f := range fields {
		switch f.Type {
		case BOOL:
			event = event.Bool(f.Key, f.Integer == 1)
		case INT:
			event = event.Int(f.Key, int(f.Integer))
		case INT64:
			event = event.Int64(f.Key, f.Integer)
		case STRING:
			event = event.Str(f.Key, f.String)
		case FLOAT:
			event = event.Float64(f.Key, f.Value.(float64))
		case ERROR:
			event = event.Err(f.Value.(error))
		case UNKNOWN:
			event = event.Any(f.Key, f.Value)
		}
	}

	return event
}

func zerologContextFields(cx zerolog.Context, fields ...Field) zerolog.Context {
	for _, f := range fields {
		switch f.Type {
		case BOOL:
			cx = cx.Bool(f.Key, f.Integer == 1)
		case INT:
			cx = cx.Int(f.Key, int(f.Integer))
		case INT64:
			cx = cx.Int64(f.Key, f.Integer)
		case STRING:
			cx = cx.Str(f.Key, f.String)
		case FLOAT:
			cx = cx.Float64(f.Key, f.Value.(float64))
		case ERROR:
			cx = cx.Err(f.Value.(error))
		case UNKNOWN:
			// do nothing
		}
	}
	return cx
}

func (z *zeroLogger) Info(msg string, fields ...Field) {
	event := z.logger.Info()
	event = zerologEventFields(event, fields...)
	event.Msg(msg)
}

func (z *zeroLogger) Debug(msg string, fields ...Field) {
	event := z.logger.Debug()
	event = zerologEventFields(event, fields...)
	event.Msg(msg)
}

func (z *zeroLogger) Warn(msg string, fields ...Field) {
	event := z.logger.Warn()
	event = zerologEventFields(event, fields...)
	event.Msg(msg)
}

func (z *zeroLogger) Error(msg string, fields ...Field) {
	event := z.logger.Error()

	if z.withStack {
		event = event.Stack()
	}

	event = zerologEventFields(event, fields...)
	event.Msg(msg)
}

func (z *zeroLogger) Panic(msg string, fields ...Field) {
	event := z.logger.Panic()
	event = zerologEventFields(event, fields...)
	event.Msg(msg)
}

func (z *zeroLogger) Fatal(msg string, fields ...Field) {
	event := z.logger.Fatal()
	event = zerologEventFields(event, fields...)
	event.Msg(msg)
}

func (z *zeroLogger) Infof(msg string, vals ...interface{}) {
	z.logger.Info().Msgf(msg, vals...)
}

func (z *zeroLogger) Errorf(msg string, vals ...interface{}) {
	event := z.logger.Error()

	if z.withStack {
		event = event.Stack()
	}

	event.Msgf(msg, vals...)
}
func (z *zeroLogger) Debugf(msg string, vals ...interface{}) {
	z.logger.Error().Msgf(msg, vals...)
}

func (z *zeroLogger) Flush() error { return nil }

func (z *zeroLogger) With(fields ...Field) Logger {
	cx := z.logger.With()
	cx = zerologContextFields(cx, fields...)
	sublogger := cx.Logger()
	return &zeroLogger{z.withStack, sublogger}
}

func (z *zeroLogger) WithContext(ctx context.Context) context.Context {
	if _, ok := ctx.Value(ctxKey{}).(Logger); !ok &&
		z.logger.GetLevel() == zerolog.Disabled {
		// Do not store disabled logger.
		return ctx
	}
	return context.WithValue(ctx, ctxKey{}, &z)
}

func (z *zeroLogger) Log(keyvals ...interface{}) error {
	z.logger.Print(keyvals...)
	return nil
}

func ZeroLoggerWithLevel(level string) ZeroLoggerOption {
	return func(zl *zeroLoggerConfig) (err error) {
		var zlevel zerolog.Level

		switch level {
		case "info":
			zlevel = zerolog.InfoLevel
		case "error":
			zlevel = zerolog.ErrorLevel
		case "debug":
			zlevel = zerolog.DebugLevel
		case "warn":
			zlevel = zerolog.WarnLevel
		default:
			zlevel = zerolog.ErrorLevel
		}

		zl.level = zlevel
		return
	}
}

func ZeroLoggerWithOutputs(outputs []string) ZeroLoggerOption {
	// /path/to/file, "stdout", "stderr"
	var writers []io.Writer

	return func(zlc *zeroLoggerConfig) (err error) {
		for _, o := range outputs {
			switch o {
			case "stdout":
				writers = append(writers, os.Stdout)
			case "stderr":
				writers = append(writers, os.Stderr)
			default:
				f, err := filepath.Abs(o)
				if err != nil {
					return fmt.Errorf("failed to find path: %s - err: %w", o, err)
				}

				_, err = os.Stat(f)
				if err != nil {
					return fmt.Errorf("failed to find path: %s - err: %w", o, err)
				}

				fd, err := os.Open(f)
				if err != nil {
					return fmt.Errorf("failed to open: %s - err: %w", o, err)
				}

				writers = append(writers, fd)
			}
		}

		if len(writers) == 0 {
			// no writer added
			writers = append(writers, os.Stdout)
		}
		zlc.writer = io.MultiWriter(writers...)
		return
	}
}

// Alerter is used to notify, how many messages were dropped
//
//	example:
//	func(missed int) {
//		fmt.Printf("Logger Dropped %d messages", missed)
//	}
type ZeroLoggerAlerter func(missed int)

func ZeroLoggerWithAsyncSink(
	size int,
	pollInterval time.Duration,
	alerter ZeroLoggerAlerter,
) ZeroLoggerOption {
	return func(zl *zeroLoggerConfig) error {
		if zl.writer == nil {
			return errors.New("no writer set, configure a writer first")
		}

		w := diode.NewWriter(
			zl.writer,
			size,
			pollInterval,
			diode.Alerter(alerter),
		)

		zl.writer = w
		return nil
	}
}

func ZeroLoggerWithCaller() ZeroLoggerOption {
	return func(zl *zeroLoggerConfig) (err error) {
		zl.withCaller = true
		return
	}
}

func ZeroLoggerWithStackTraceInError() ZeroLoggerOption {
	return func(zlc *zeroLoggerConfig) (err error) {
		zlc.withStack = true
		return
	}
}

func ZeroLoggerWithFields(fields ...Field) ZeroLoggerOption {
	return func(zlc *zeroLoggerConfig) (err error) {
		zlc.fields = append(zlc.fields, fields...)
		return
	}
}

func (zlc *zeroLoggerConfig) build() (Logger, error) {
	zlg := zerolog.New(zlc.writer)
	zlg = zlg.Level(zlc.level)

	if zlc.withCaller {
		zlg = zlg.With().Caller().Logger()
	}

	if len(zlc.fields) > 0 {
		cx := zlg.With()
		cx = zerologContextFields(cx, zlc.fields...)
		zlg = cx.Logger()
	}

	return &zeroLogger{zlc.withStack, zlg}, nil
}

func NewZeroLogger(options ...ZeroLoggerOption) (Logger, error) {
	cfg := &zeroLoggerConfig{
		level:      zerolog.DebugLevel,
		writer:     os.Stdout,
		withCaller: false,
		withStack:  false,
		fields:     []Field{String("logger", "zerolog")},
	}

	for _, ofn := range options {
		err := ofn(cfg)
		if err != nil {
			return nil, err
		}
	}

	return cfg.build()
}
