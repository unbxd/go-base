package base

import (
	"github.com/uknth/go-base/base/log"
)

// Base is a gener
type Base struct {
	logger log.Logger
}

// NewBase returns the Base structure
func NewBase(
	logger log.Logger,
) (Base, error) {
	return Base{}, nil
}
