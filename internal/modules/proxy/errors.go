package proxy

import "errors"

var (
	ErrStartingServer    = errors.New("error starting server")
	ErrTimeoutReached    = errors.New("timeout reached")
	ErrCannotSwitchState = errors.New("cannot switch state")
)
