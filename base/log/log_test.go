package log

import (
	"errors"
	"reflect"
	"testing"
)

func TestString(t *testing.T) {
	type args struct {
		key   string
		value string
	}
	tests := []struct {
		name string
		args args
		want Field
	}{
		{"random string", args{"message", "some-message"}, Field{Key: "message", Type: STRING, String: "some-message"}},
		{"numbered string", args{"123123", "123123"}, Field{Key: "123123", Type: STRING, String: "123123"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := String(tt.args.key, tt.args.value); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInt(t *testing.T) {
	type args struct {
		key   string
		value int
	}
	tests := []struct {
		name string
		args args
		want Field
	}{
		{"integers", args{"random-int", 123}, Field{Key: "random-int", Type: INT, Integer: 123}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Int(tt.args.key, tt.args.value); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Int() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInt64(t *testing.T) {
	type args struct {
		key   string
		value int64
	}
	tests := []struct {
		name string
		args args
		want Field
	}{
		{"integers", args{"random-int", 123}, Field{Key: "random-int", Type: INT, Integer: 123}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Int64(tt.args.key, tt.args.value); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Int64() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBool(t *testing.T) {
	type args struct {
		key   string
		value bool
	}
	tests := []struct {
		name string
		args args
		want Field
	}{
		{"true", args{"key", true}, Field{Key: "key", Type: BOOL, Integer: 1}},
		{"false", args{"key", false}, Field{Key: "key", Type: BOOL, Integer: 0}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Bool(tt.args.key, tt.args.value); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Bool() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestError(t *testing.T) {
	type args struct {
		err error
	}
	tests := []struct {
		name string
		args args
		want Field
	}{
		// TODO: Add test cases.
		{"has-error", args{errors.New("some-error")}, Field{Key: "err", Type: ERROR, Value: errors.New("some-error")}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Error(tt.args.err); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Error() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFloat(t *testing.T) {
	type args struct {
		key   string
		value float64
	}
	tests := []struct {
		name string
		args args
		want Field
	}{
		{"floating-point", args{"key", 0.123}, Field{Key: "key", Type: FLOAT, Value: 0.123}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Float(tt.args.key, tt.args.value); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Float() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReflect(t *testing.T) {
	type args struct {
		key   string
		value interface{}
	}
	tests := []struct {
		name string
		args args
		want Field
	}{
		{"map", args{"key", make(map[string]string)}, Field{Key: "key", Type: UNKNOWN, Value: make(map[string]string)}},
		{"error", args{"key", errors.New("some-error")}, Field{Key: "key", Type: UNKNOWN, Value: errors.New("some-error")}},
		{"number", args{"key", 123}, Field{Key: "key", Type: UNKNOWN, Value: 123}},
		{"error", args{"key", "some-string"}, Field{Key: "key", Type: UNKNOWN, Value: "some-string"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Reflect(tt.args.key, tt.args.value); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Reflect() = %v, want %v", got, tt.want)
			}
		})
	}
}
