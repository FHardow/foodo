package webpush

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	sherwebpush "github.com/SherClockHolmes/webpush-go"
	"github.com/fhardow/foodo/internal/domain/order"
	"github.com/fhardow/foodo/internal/domain/push"
	"github.com/fhardow/foodo/internal/testutil/mock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestOrder(t *testing.T, userID uuid.UUID) *order.Order {
	t.Helper()
	o, err := order.New(userID)
	require.NoError(t, err)
	return o
}

func seedSubscription(t *testing.T, repo *mock.PushRepo, userID uuid.UUID, endpoint string) {
	t.Helper()
	sub, err := push.New(userID, endpoint, "p256dh-key", "auth-key")
	require.NoError(t, err)
	require.NoError(t, repo.Save(context.Background(), sub))
}

func TestOrderAccepted_SendsExpectedPayloadToEachSubscription(t *testing.T) {
	repo := mock.NewPushRepo()
	userID := uuid.New()
	seedSubscription(t, repo, userID, "https://push.example.com/1")
	seedSubscription(t, repo, userID, "https://push.example.com/2")

	var sentMessages [][]byte
	var sentEndpoints []string
	n := &Notifier{
		subs: repo,
		send: func(message []byte, s *sherwebpush.Subscription, options *sherwebpush.Options) (*http.Response, error) {
			sentMessages = append(sentMessages, message)
			sentEndpoints = append(sentEndpoints, s.Endpoint)
			return &http.Response{StatusCode: http.StatusCreated, Body: http.NoBody}, nil
		},
		vapidPub:  "pub",
		vapidPriv: "priv",
		subject:   "mailto:owner@example.com",
	}

	o := newTestOrder(t, userID)
	n.OrderAccepted(o)

	require.Len(t, sentMessages, 2)
	assert.ElementsMatch(t, []string{"https://push.example.com/1", "https://push.example.com/2"}, sentEndpoints)

	var payload notificationPayload
	require.NoError(t, json.Unmarshal(sentMessages[0], &payload))
	assert.Equal(t, "Order accepted", payload.Title)
	assert.Equal(t, "Your order has been accepted.", payload.Body)
	assert.Equal(t, "/orders/"+o.ID().String(), payload.URL)
}

func TestOrderStarted_SendsExpectedPayload(t *testing.T) {
	repo := mock.NewPushRepo()
	userID := uuid.New()
	seedSubscription(t, repo, userID, "https://push.example.com/1")

	var sent notificationPayload
	n := &Notifier{
		subs: repo,
		send: func(message []byte, s *sherwebpush.Subscription, options *sherwebpush.Options) (*http.Response, error) {
			require.NoError(t, json.Unmarshal(message, &sent))
			return &http.Response{StatusCode: http.StatusCreated, Body: http.NoBody}, nil
		},
	}

	n.OrderStarted(newTestOrder(t, userID))

	assert.Equal(t, "Being prepared", sent.Title)
	assert.Equal(t, "Your order is being prepared.", sent.Body)
}

func TestOrderFinished_SendsExpectedPayload(t *testing.T) {
	repo := mock.NewPushRepo()
	userID := uuid.New()
	seedSubscription(t, repo, userID, "https://push.example.com/1")

	var sent notificationPayload
	n := &Notifier{
		subs: repo,
		send: func(message []byte, s *sherwebpush.Subscription, options *sherwebpush.Options) (*http.Response, error) {
			require.NoError(t, json.Unmarshal(message, &sent))
			return &http.Response{StatusCode: http.StatusCreated, Body: http.NoBody}, nil
		},
	}

	n.OrderFinished(newTestOrder(t, userID))

	assert.Equal(t, "Order ready", sent.Title)
	assert.Equal(t, "Your order is ready!", sent.Body)
}

func TestNotify_DeletesSubscriptionOnGone(t *testing.T) {
	repo := mock.NewPushRepo()
	userID := uuid.New()
	seedSubscription(t, repo, userID, "https://push.example.com/gone")

	n := &Notifier{
		subs: repo,
		send: func(message []byte, s *sherwebpush.Subscription, options *sherwebpush.Options) (*http.Response, error) {
			return &http.Response{StatusCode: http.StatusGone, Body: http.NoBody}, nil
		},
	}

	n.OrderAccepted(newTestOrder(t, userID))

	remaining, err := repo.ListByUser(context.Background(), userID)
	require.NoError(t, err)
	assert.Empty(t, remaining, "subscription must be deleted after a 410 Gone response")
}

func TestNotify_SendErrorDoesNotPanic(t *testing.T) {
	repo := mock.NewPushRepo()
	userID := uuid.New()
	seedSubscription(t, repo, userID, "https://push.example.com/broken")

	n := &Notifier{
		subs: repo,
		send: func(message []byte, s *sherwebpush.Subscription, options *sherwebpush.Options) (*http.Response, error) {
			return nil, errors.New("network unreachable")
		},
	}

	assert.NotPanics(t, func() { n.OrderAccepted(newTestOrder(t, userID)) })
}
