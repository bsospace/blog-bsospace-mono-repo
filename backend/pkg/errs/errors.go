package errs

import "errors"

// Shared business errors
var (
	ErrPostNotFound   = errors.New("post not found")
	ErrUnauthorized   = errors.New("unauthorized access")
	ErrInvalidPayload = errors.New("invalid payload")
)
