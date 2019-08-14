package log

import (
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type zapLogger struct {
	zapLogger *zap.Logger
}

func (zl *zapLogger) convert(fields ...Field) []zap.Field {
	var zfields []zap.Field
	for _, fl := range fields {
		switch fl.Type {
		case BOOL:
			var bl bool
			if fl.Integer == 1 {
				bl = true
			}
			zfields = append(zfields, zap.Bool(fl.Key, bl))
		case ERROR:
			zfields = append(zfields, zap.Error(fl.Value.(error)))
		case FLOAT:
			zfields = append(zfields, zap.Float64(fl.Key, fl.Value.(float64)))
		case INT:
			zfields = append(zfields, zap.Int64(fl.Key, fl.Integer))
		case STRING:
			zfields = append(zfields, zap.String(fl.Key, fl.String))
		case UNKNOWN:
			zfields = append(zfields, zap.Any(fl.Key, fl.Value))
		default:
			zfields = append(zfields, zap.Any(fl.Key, fl.Value))
		}
	}
	return zfields
}

func (zl *zapLogger) Info(msg string, fields ...Field) {
	zl.zapLogger.Info(msg, zl.convert(fields...)...)
}

func (zl *zapLogger) Warn(msg string, fields ...Field) {
	zl.zapLogger.Warn(msg, zl.convert(fields...)...)
}

func (zl *zapLogger) Error(msg string, fields ...Field) {
	zl.zapLogger.Error(msg, zl.convert(fields...)...)
}

func (zl *zapLogger) Panic(msg string, fields ...Field) {
	zl.zapLogger.Panic(msg, zl.convert(fields...)...)
}

func (zl *zapLogger) Fatal(msg string, fields ...Field) {
	zl.zapLogger.Fatal(msg, zl.convert(fields...)...)
}

func (zl *zapLogger) Flush() error {
	return zl.zapLogger.Sync()
}

func (zl *zapLogger) With(fields ...Field) Logger {
	zl.zapLogger = zl.zapLogger.With(zl.convert(fields...)...)
	return zl
}

func (zl *zapLogger) Log(kv ...interface{}) error {
	zl.zapLogger.Sugar().Infow("", kv...)
	return nil
}

func zapLevel(level string) zap.AtomicLevel {
	switch level {
	case "info":
		return zap.NewAtomicLevelAt(zap.InfoLevel)
	case "error":
		return zap.NewAtomicLevelAt(zap.ErrorLevel)
	case "debug":
		return zap.NewAtomicLevelAt(zap.DebugLevel)
	case "warn":
		return zap.NewAtomicLevelAt(zap.WarnLevel)
	default:
		return zap.NewAtomicLevelAt(zap.ErrorLevel)
	}
}

// NewZapLogger returns a default implementaiton of log.Logger interface
// using uber/zap as core layer
func NewZapLogger(
	level string,
	encoding string,
	output []string,
) (Logger, error) {
	config := zap.NewProductionConfig()

	config.Level = zapLevel(level)
	config.OutputPaths = output
	config.Encoding = encoding

	zl, err := config.Build()
	if err != nil {
		return nil, errors.Wrap(err, "unable to build zap logger")
	}
	return &zapLogger{zl}, err
}
