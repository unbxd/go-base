package errors

import (
	builtin_errors "errors"
	"fmt"
)

// With is easy error access
func With(err error, errors ...error) error {
	return builtin_errors.Join(append([]error{err}, errors...)...)
}

// Builtin Methods for Errors Package
func Is(err, target error) bool     { return builtin_errors.Is(err, target) }
func As(err error, target any) bool { return builtin_errors.As(err, target) }
func Join(errors ...error) error    { return builtin_errors.Join(errors...) }
func Unwrap(err error) error        { return builtin_errors.Unwrap(err) }
func New(msg string) error          { return builtin_errors.New(msg) }

// Method from github.com/pkg/errors
func Wrap(err error, str string) error { return fmt.Errorf(str+": [%w]", err) }
func Cause(err error) error            { return builtin_errors.Unwrap(err) }
func Wrapf(err error, fmtstr string, args ...interface{}) error {
	return fmt.Errorf(fmt.Sprintf(fmtstr, args...)+": [%w]", err)
}
