package order

// CustomerNotifier is called after order transitions the customer should be
// told about. Unlike Notifier (admin-facing, Telegram), this fires only on
// forward status progress: accepted, started, finished.
type CustomerNotifier interface {
	OrderAccepted(o *Order)
	OrderStarted(o *Order)
	OrderFinished(o *Order)
}

type noopCustomerNotifier struct{}

func (noopCustomerNotifier) OrderAccepted(*Order) {}
func (noopCustomerNotifier) OrderStarted(*Order)  {}
func (noopCustomerNotifier) OrderFinished(*Order) {}
