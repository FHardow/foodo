package push

import (
	"time"

	domerrors "github.com/fhardow/foodo/pkg/errors"
	"github.com/google/uuid"
)

type ID = uuid.UUID

// Subscription is one browser+device push registration for a customer.
type Subscription struct {
	id        ID
	userID    uuid.UUID
	endpoint  string
	p256dh    string
	auth      string
	createdAt time.Time
}

func New(userID uuid.UUID, endpoint, p256dh, auth string) (*Subscription, error) {
	if userID == uuid.Nil {
		return nil, domerrors.BadRequest("user ID is required")
	}
	if endpoint == "" {
		return nil, domerrors.BadRequest("endpoint is required")
	}
	if p256dh == "" {
		return nil, domerrors.BadRequest("p256dh is required")
	}
	if auth == "" {
		return nil, domerrors.BadRequest("auth is required")
	}
	return &Subscription{
		id:        uuid.New(),
		userID:    userID,
		endpoint:  endpoint,
		p256dh:    p256dh,
		auth:      auth,
		createdAt: time.Now().UTC(),
	}, nil
}

// Reconstitute rebuilds a Subscription from persistence without re-running validation.
func Reconstitute(id, userID uuid.UUID, endpoint, p256dh, auth string, createdAt time.Time) *Subscription {
	return &Subscription{
		id:        id,
		userID:    userID,
		endpoint:  endpoint,
		p256dh:    p256dh,
		auth:      auth,
		createdAt: createdAt,
	}
}

func (s *Subscription) ID() ID               { return s.id }
func (s *Subscription) UserID() uuid.UUID    { return s.userID }
func (s *Subscription) Endpoint() string     { return s.endpoint }
func (s *Subscription) P256dh() string       { return s.p256dh }
func (s *Subscription) Auth() string         { return s.auth }
func (s *Subscription) CreatedAt() time.Time { return s.createdAt }
