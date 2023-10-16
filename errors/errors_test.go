package errors

import (
	berrs "errors"
	"testing"
)

func TestWrap(t *testing.T) {
	type args struct {
		err error
		str string
	}

	tests := []struct {
		name string
		args args
		want string
	}{
		{"simple wrap", args{berrs.New("error1"), "outer string"}, "outer string: [error1]"},
		{"special characters", args{berrs.New("error1"), "error string with :"}, "error string with :: [error1]"},
		{"empty string", args{berrs.New("error1"), ""}, ": [error1]"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Wrap(tt.args.err, tt.args.str); err.Error() != tt.want {
				t.Errorf("Wrap() error = %v, wantErr %v", err, tt.want)
			}
		})
	}
}

func TestWrapWithIs(t *testing.T) {
	var (
		err1 = berrs.New("error 1")
		err2 = berrs.New("error 2")
		err3 = berrs.New("error 3")
	)

	type args struct {
		err error
		str string
	}

	tests := []struct {
		name string
		args args
		want error
	}{
		{"simple wrap", args{err1, "outer string"}, err1},
		{"special characters", args{err2, "string with colon :"}, err2},
		{"empty string", args{err3, ""}, err3},
	}

	for _, tt := range tests {
		t.Run(
			tt.name+"-Is", func(t *testing.T) {
				err := Wrap(tt.args.err, tt.args.str)

				if !Is(err, tt.want) {
					t.Errorf("Wrap() error = %v, wantErr %v", err, tt.want)
				}

				if Cause(err) != tt.args.err {
					t.Errorf("Cause() error = %v, wantErr %v", err, tt.args.err)
				}

				if e := Unwrap(err); e != tt.args.err {
					t.Errorf("Unwrap() error = %v, wantErr %v", err, tt.args.err)
				}
			},
		)
	}

	for _, tt := range tests {
		t.Run(tt.name+"-Unwrap", func(t *testing.T) {
			err := Wrap(tt.args.err, tt.args.str)

			ecause := Cause(err)
			eunwrap := Unwrap(err)

			if ecause != tt.want {
				t.Errorf("Wrap() error = %v, wantErr %v", ecause, tt.want)
			}
			if eunwrap != tt.want {
				t.Errorf("Wrap() error = %v, wantErr %v", eunwrap, tt.want)
			}

			if ecause != eunwrap {
				t.Errorf("Unwrap & Cause are same err_cause = %v, err_unwrap = %v ", ecause, eunwrap)
			}
		})
	}

}

func TestWith(t *testing.T) {
	e1 := New("err 1")
	e2 := New("err 2")
	e3 := New("err 3")
	e4 := New("err 4")
	e5 := New("err 5")

	type args struct {
		err  error
		errs []error
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		is      error
		cause   error
	}{
		// TODO: Add test cases.
		{"join test", args{e1, []error{e2, e3}}, true, e2, e1},
		{"join test 2", args{e1, []error{e4, e5}}, true, e5, e1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := With(tt.args.err, tt.args.errs...)
			if (err != nil) != tt.wantErr {
				t.Errorf("With() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !Is(err, tt.is) {
				t.Errorf("Is() error = %v, wantErr %v", err, tt.is)
			}

			ec := Cause(err)
			if ec != tt.cause {
				t.Errorf("Cause() error = %v, wantErr %v", err, tt.cause)
			}

		})
	}
}
