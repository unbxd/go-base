package notifier

import "context"

type noopNotifier struct{}

func (nn *noopNotifier) Notify(
	cx context.Context,
	data interface{},
) error {
	return nil
}

func NewNoopNotifier() Notifier { return &noopNotifier{} }
