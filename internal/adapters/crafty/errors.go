package crafty

import "errors"

var (
	ErrHTTPRequestFailed   = errors.New("failed to send http request")
	ErrFailedToReadBody    = errors.New("failed to read response body")
	ErrFailedToGetServers  = errors.New("failed to get servers")
	ErrFailedToStartServer = errors.New("failed to start MC server")
	ErrFailedToStopServer  = errors.New("failed to stop MC server")
	ErrAuthorizationFailed = errors.New("authorization failed")
	ErrNoSuchServer        = errors.New("no such server")
)
