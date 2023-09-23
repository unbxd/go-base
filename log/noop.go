package log

import "context"

type noopLogger struct{}

func (nl *noopLogger) Info(string, ...Field)                          {}
func (nl *noopLogger) Warn(string, ...Field)                          {}
func (nl *noopLogger) Error(string, ...Field)                         {}
func (nl *noopLogger) Panic(string, ...Field)                         {}
func (nl *noopLogger) Fatal(string, ...Field)                         {}
func (nl *noopLogger) Debug(string, ...Field)                         {}
func (nl *noopLogger) Infof(string, ...interface{})                   {}
func (nl *noopLogger) Errorf(string, ...interface{})                  {}
func (nl *noopLogger) Debugf(string, ...interface{})                  {}
func (nl *noopLogger) Flush() error                                   { return nil }
func (nl *noopLogger) With(...Field) Logger                           { return &noopLogger{} }
func (nl *noopLogger) WithContext(cx context.Context) context.Context { return cx }
func (nl *noopLogger) Log(keyvals ...interface{}) error               { return nil }

func NewNoopLogger() Logger { return &noopLogger{} }
