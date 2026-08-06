package webpush

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	sherwebpush "github.com/SherClockHolmes/webpush-go"
	"github.com/fhardow/foodo/internal/domain/order"
	"github.com/fhardow/foodo/internal/domain/push"
)

type sendFunc func(message []byte, s *sherwebpush.Subscription, options *sherwebpush.Options) (*http.Response, error)

type Notifier struct {
	subs      push.Repository
	send      sendFunc
	vapidPub  string
	vapidPriv string
	subject   string
}

var _ order.CustomerNotifier = (*Notifier)(nil)

func NewNotifier(subs push.Repository, vapidPub, vapidPriv, subject string) *Notifier {
	return &Notifier{
		subs:      subs,
		send:      sherwebpush.SendNotification,
		vapidPub:  vapidPub,
		vapidPriv: vapidPriv,
		subject:   subject,
	}
}

type notificationPayload struct {
	Title string `json:"title"`
	Body  string `json:"body"`
	URL   string `json:"url"`
}

func (n *Notifier) OrderAccepted(o *order.Order) {
	n.notify(o, "Order accepted", "Your order has been accepted.")
}

func (n *Notifier) OrderStarted(o *order.Order) {
	n.notify(o, "Being prepared", "Your order is being prepared.")
}

func (n *Notifier) OrderFinished(o *order.Order) {
	n.notify(o, "Order ready", "Your order is ready!")
}

func (n *Notifier) notify(o *order.Order, title, body string) {
	ctx := context.Background()
	subs, err := n.subs.ListByUser(ctx, o.UserID())
	if err != nil {
		slog.Error("push: failed to list subscriptions", "err", err, "order_id", o.ID())
		return
	}
	msg, err := json.Marshal(notificationPayload{Title: title, Body: body, URL: "/orders/" + o.ID().String()})
	if err != nil {
		slog.Error("push: failed to marshal payload", "err", err)
		return
	}
	for _, s := range subs {
		resp, err := n.send(msg, &sherwebpush.Subscription{
			Endpoint: s.Endpoint(),
			Keys:     sherwebpush.Keys{P256dh: s.P256dh(), Auth: s.Auth()},
		}, &sherwebpush.Options{
			Subscriber:      n.subject,
			VAPIDPublicKey:  n.vapidPub,
			VAPIDPrivateKey: n.vapidPriv,
			TTL:             30,
		})
		if err != nil {
			slog.Error("push: send failed", "err", err, "order_id", o.ID())
			continue
		}
		resp.Body.Close()
		if resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusGone {
			if delErr := n.subs.DeleteByEndpoint(ctx, s.Endpoint()); delErr != nil {
				slog.Error("push: failed to delete stale subscription", "err", delErr)
			}
		}
	}
}
