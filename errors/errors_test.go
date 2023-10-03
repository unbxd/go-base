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
