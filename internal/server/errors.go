// Package server defines the standard error type for the server.
package server

import (
	"fmt"
)

type ServerError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Error implements the error interface for ServerError.
func (e *ServerError) Error() string {
	return fmt.Sprintf("code %d: %s", e.Code, e.Message)
}