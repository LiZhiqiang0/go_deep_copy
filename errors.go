package go_deep_copy

import "errors"

var (
	ErrInvalidCopyDestination = errors.New("copy destination must be non-nil and addressable")
	ErrInvalidCopyFrom        = errors.New("copy from must be non-nil and addressable")
	ErrNotSupported           = errors.New("not supported")
)
