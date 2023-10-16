package errors

import (
	builtin_errors "errors"
	"fmt"
)

type causer interface {
	Cause() error
}

type withCause struct {
	cause error
	msg   string
}

func (wc *withCause) Cause() error  { return wc.cause }
func (wc *withCause) Unwrap() error { return wc.cause }
func (wc *withCause) Error() string { return fmt.Sprintf("%s: %s", wc.msg, wc.cause.Error()) }

// With is easy error access
func With(err error, errs ...error) error {
	return WithMessage(err, "causes", errs...)
}

func WithMessage(err error, msg string, errs ...error) error {
	return builtin_errors.Join(
		append(
			[]error{&withCause{
				err, msg,
			}},
			errs...,
		)...,
	)
}

// Builtin Methods for Errors Package
func Is(err, target error) bool { return builtin_errors.Is(err, target) }

// TODO: As test cases
func As(err error, target any) bool { return builtin_errors.As(err, target) }
func Join(errors ...error) error    { return builtin_errors.Join(errors...) }
func Unwrap(err error) error        { return builtin_errors.Unwrap(err) }
func New(msg string) error          { return builtin_errors.New(msg) }

// Method from github.com/unbxd/go-base/v2/errors
func Wrap(err error, message string) error {
	if err == nil {
		return nil
	}

	return &withCause{
		err, message,
	}
}

func Cause(err error) error {
	if eu, ok := err.(interface{ Unwrap() []error }); ok {
		var eret error

		errs := eu.Unwrap()

		for _, er := range errs {
			c, ok := er.(causer)
			if ok {
				eret = c.Cause()
				break
			}
		}

		if eret == nil {
			eret = errs[len(errs)-1]
		}

		return eret
	} else {
		for err != nil {
			cause, ok := err.(causer)
			if !ok {
				break
			}
			err = cause.Cause()
		}
	}
	return err
}
func Wrapf(err error, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}

	return &withCause{
		err, fmt.Sprintf(format, args...),
	}
}
