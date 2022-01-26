package client

import (
	"errors"
	"fmt"
)

// errConnectionFailed implements an error returned when connection failed.
type errConnectionFailed struct {
	host string
}

// Error returns a string representation of an errConnectionFailed
func (err errConnectionFailed) Error() string {
	if err.host == "" {
		return "Cannot connect to the agent. Is the agent running on this host?"
	}
	return fmt.Sprintf("Cannot connect to the agent at %s. Is the agent running?", err.host)
}

// IsErrConnectionFailed returns true if the error is caused by connection failed.
func IsErrConnectionFailed(err error) bool {
	var connErr *errConnectionFailed
	return errors.As(err, &connErr)
}

// ErrorConnectionFailed returns an error with host in the error message when connection to agent failed.
func ErrorConnectionFailed(host string) error {
	return errConnectionFailed{host: host}
}
