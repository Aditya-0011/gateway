package timeout

import (
	"context"
	"time"
)

const (
	Duration = 9500 * time.Millisecond
)

func WithDeadline(c context.Context, d time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(c, d)
}
