package auth

import "errors"

var (
	// ErrForbidden occurs when user is identified but not authorized
	ErrForbidden = errors.New("forbidden access")
)
