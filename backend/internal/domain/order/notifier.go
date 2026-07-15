package order

// Notifier is called after significant order state transitions.
type Notifier interface {
	OrderConfirmed(o *Order)
}

type noopNotifier struct{}

func (noopNotifier) OrderConfirmed(*Order) {}
