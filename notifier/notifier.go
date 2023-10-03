package notifier

import (
	"context"
)

type (
	Notifier interface {
		// Notify publishes data to the notifier
		// after decorating it with additional details
		Notify(
			cx context.Context, data interface{},
		) error
	}
)
