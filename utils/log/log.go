package log

import (
	kit_log "github.com/go-kit/log"
)

// FieldType defines the type for a field
type FieldType int

// Field Types supported by Logger
const (
	UNKNOWN FieldType = iota
	BOOL
	INT
	STRING
	ERROR
	FLOAT
)

// Field defines a standard Key-Value pair used to populate
// the value for the logger
type Field struct {
	// Key for a given field
	Key string
	// Type of a given field
	Type FieldType
	// Value of the given field
	Value interface{}
	// optmizations for string & int
	Integer int64
	String  string
}

// String is logger wrapper for string
func String(key string, value string) Field {
	return Field{Key: key, Type: STRING, String: value}
}

// Int wrapper for logging
func Int(key string, value int) Field {
	return Field{Key: key, Type: INT, Integer: int64(value)}
}

// Int64 is a wrapper int64 values for logging
func Int64(key string, value int64) Field {
	return Field{Key: key, Type: INT, Integer: int64(value)}
}

// Bool is a wrapper for boolean values for logging
func Bool(key string, value bool) Field {
	var ival int

	if value {
		ival = 1
	}

	return Field{Key: key, Type: BOOL, Integer: int64(ival)}
}

// Error is wrapper for error values for logging
func Error(err error) Field {
	return Field{Key: "err", Type: ERROR, Value: err}
}

// Float value is for floating point fields
func Float(key string, value float64) Field {
	return Field{Key: key, Type: FLOAT, Value: value}
}

// Reflect returns a field for which the value is undetermined
func Reflect(key string, value interface{}) Field {
	return Field{Key: key, Type: UNKNOWN, Value: value}
}

// Logger defines standard set of functions supported by logger and
// used across the Base Package
type Logger interface {
	// Should also implement default kit.log.Logger interface
	kit_log.Logger

	// Leveled Logging
	Info(string, ...Field)
	Warn(string, ...Field)
	Error(string, ...Field)
	Panic(string, ...Field)
	Fatal(string, ...Field)
	Debug(string, ...Field)

	Infof(string, ...interface{})
	Errorf(string, ...interface{})
	Debugf(string, ...interface{})

	// Flush the logs
	Flush() error

	// Default Fields
	With(...Field) Logger
}
