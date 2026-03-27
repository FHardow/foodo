package order

import "context"

// Notifier sends notifications about order lifecycle events.
type Notifier interface {
	OrderCreated(ctx context.Context, o *Order) error
}
