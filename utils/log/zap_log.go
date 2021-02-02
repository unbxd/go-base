package log

import (
	"go.uber.org/zap"
)

type (
	ZapLoggerOption func(*zapLogger)

	zapLogger struct {
		config    *zap.Config
		zapLogger *zap.Logger
	}
)

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
			if fl.Value != nil {
				zfields = append(zfields, zap.Error(fl.Value.(error)))
			} else {
				zfields = append(zfields, zap.Error(nil))
			}
		case FLOAT:
			if fl.Value != nil {
				zfields = append(zfields, zap.Float64(fl.Key, fl.Value.(float64)))
			}
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

func (zl *zapLogger) Debug(msg string, fields ...Field) {
	zl.zapLogger.Debug(msg, zl.convert(fields...)...)
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

func (zl *zapLogger) Infof(msg string, vals ...interface{}) {
	zl.zapLogger.Sugar().Infof(msg, vals...)
}
func (zl *zapLogger) Errorf(msg string, vals ...interface{}) {
	zl.zapLogger.Sugar().Errorf(msg, vals...)
}
func (zl *zapLogger) Debugf(msg string, vals ...interface{}) {
	zl.zapLogger.Sugar().Debugf(msg, vals...)
}

func (zl *zapLogger) Flush() error {
	return zl.zapLogger.Sync()
}

func (zl *zapLogger) With(fields ...Field) Logger {
	if len(fields) == 0 {
		return zl
	}
	l := zl.clone()
	l.zapLogger = l.zapLogger.With(zl.convert(fields...)...)
	return l
}

func (zl *zapLogger) clone() *zapLogger {
	copy := *zl
	return &copy
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

var (
	defaultLevel    = zap.NewAtomicLevelAt(zap.ErrorLevel)
	defaultOutputs  = []string{"stdout"}
	defaultEncoding = "json"
)

// ZapWithLevel is option to set level for Zap Based Logger
func ZapWithLevel(level string) ZapLoggerOption {
	return func(zl *zapLogger) {
		zl.config.Level = zapLevel(level)
	}
}

//ZapWithEncoding is option to set encoding for zap based logger
func ZapWithEncoding(encoding string) ZapLoggerOption {
	return func(zl *zapLogger) {
		zl.config.Encoding = encoding
	}
}

// ZapWithOutput is option to set output paths for zap based logger
func ZapWithOutput(outputs []string) ZapLoggerOption {
	return func(zl *zapLogger) {
		zl.config.OutputPaths = outputs
	}
}

// ZapWithAppendedOutput is option to add additional output to list of
// existing output destination
func ZapWithAppendedOutput(outputs []string) ZapLoggerOption {
	return func(zl *zapLogger) {
		zl.config.OutputPaths = append(
			zl.config.OutputPaths, outputs...,
		)
	}
}

// NewZapLogger returns a default implementaiton of log.Logger interface
// using uber/zap as core layer
func NewZapLogger(
	options ...ZapLoggerOption,
) (Logger, error) {
	config := zap.NewProductionConfig()

	config.Level = defaultLevel
	config.OutputPaths = defaultOutputs
	config.Encoding = defaultEncoding

	zl := &zapLogger{config: &config}

	for _, opt := range options {
		opt(zl)
	}

	// build the logger
	logger, err := zl.config.Build(
		zap.AddCallerSkip(1),
	)
	if err != nil {
		return nil, err
	}

	zl.zapLogger = logger
	return zl, nil
}
