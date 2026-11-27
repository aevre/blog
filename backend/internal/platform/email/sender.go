// Package email provides emial sender
package email

import (
	"context"
)

// Sender defines a type that can send messages asynchronously using the provided context.
type Sender interface {
	// Send attempts to send the message asynchronously using the given context for cancellation or deadlines.
	Send(ctx context.Context, message Message) error
}
