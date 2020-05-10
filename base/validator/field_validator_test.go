package validator

import (
	"context"
	"strings"
	"testing"
)

func TestBasicValidation(t *testing.T) {
	type Base struct {
		A string   `valid:"required"`
		B []string `valid:"min=1"`
	}

	base := Base{
		A: "",
		B: nil,
	}

	v := NewFieldValidator(WithTag("valid"))
	err := v.Validate(context.Background(), base)

	if err == nil {
		t.Errorf("test failed")
	} else {
		if !strings.Contains(err.Error(), "Field validation for 'A' failed on the 'required' tag") {
			t.Errorf("failed on required tag")
		}
		if !strings.Contains(err.Error(), "Field validation for 'B' failed on the 'min' tag") {
			t.Errorf("failed on B min length")
		}
		_, ok := err.(FieldValidationError)
		if !ok {
			t.Errorf("failed on error type")
		}
	}

}
