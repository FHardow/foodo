package postgres

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"

	"github.com/fhardow/foodo/internal/domain/push"
	"github.com/fhardow/foodo/internal/infra/postgres/models"
	domerrors "github.com/fhardow/foodo/pkg/errors"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type pushSubscriptionRepo struct {
	db  *gorm.DB
	key []byte
}

// NewPushSubscriptionRepo decodes encryptionKey (base64, must be exactly 32
// bytes for AES-256) once at construction time and fails fast if it's invalid.
func NewPushSubscriptionRepo(db *gorm.DB, encryptionKey string) (push.Repository, error) {
	key, err := base64.StdEncoding.DecodeString(encryptionKey)
	if err != nil {
		return nil, fmt.Errorf("decode push encryption key: %w", err)
	}
	if len(key) != 32 {
		return nil, fmt.Errorf("push encryption key must decode to 32 bytes, got %d", len(key))
	}
	return &pushSubscriptionRepo{db: db, key: key}, nil
}

func (r *pushSubscriptionRepo) Save(ctx context.Context, s *push.Subscription) error {
	encEndpoint, err := encryptField(r.key, s.Endpoint())
	if err != nil {
		return err
	}
	encP256dh, err := encryptField(r.key, s.P256dh())
	if err != nil {
		return err
	}
	encAuth, err := encryptField(r.key, s.Auth())
	if err != nil {
		return err
	}
	m := models.PushSubscription{
		ID:           s.ID().String(),
		UserID:       s.UserID().String(),
		Endpoint:     encEndpoint,
		EndpointHash: hashEndpoint(s.Endpoint()),
		P256dh:       encP256dh,
		Auth:         encAuth,
		CreatedAt:    s.CreatedAt(),
	}
	return dbFromCtx(ctx, r.db).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "endpoint_hash"}},
		DoUpdates: clause.AssignmentColumns([]string{"user_id", "endpoint", "p256dh", "auth"}),
	}).Create(&m).Error
}

func (r *pushSubscriptionRepo) ListByUser(ctx context.Context, userID uuid.UUID) ([]*push.Subscription, error) {
	var ms []models.PushSubscription
	if err := dbFromCtx(ctx, r.db).Where("user_id = ?", userID.String()).Find(&ms).Error; err != nil {
		return nil, err
	}
	subs := make([]*push.Subscription, 0, len(ms))
	for i := range ms {
		s, err := r.toDomain(&ms[i])
		if err != nil {
			return nil, err
		}
		subs = append(subs, s)
	}
	return subs, nil
}

func (r *pushSubscriptionRepo) FindByEndpoint(ctx context.Context, endpoint string) (*push.Subscription, error) {
	var m models.PushSubscription
	err := dbFromCtx(ctx, r.db).First(&m, "endpoint_hash = ?", hashEndpoint(endpoint)).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domerrors.NotFound("push subscription not found")
		}
		return nil, err
	}
	return r.toDomain(&m)
}

func (r *pushSubscriptionRepo) DeleteByEndpoint(ctx context.Context, endpoint string) error {
	return dbFromCtx(ctx, r.db).Where("endpoint_hash = ?", hashEndpoint(endpoint)).Delete(&models.PushSubscription{}).Error
}

func (r *pushSubscriptionRepo) toDomain(m *models.PushSubscription) (*push.Subscription, error) {
	id, err := uuid.Parse(m.ID)
	if err != nil {
		return nil, err
	}
	userID, err := uuid.Parse(m.UserID)
	if err != nil {
		return nil, err
	}
	endpoint, err := decryptField(r.key, m.Endpoint)
	if err != nil {
		return nil, err
	}
	p256dh, err := decryptField(r.key, m.P256dh)
	if err != nil {
		return nil, err
	}
	auth, err := decryptField(r.key, m.Auth)
	if err != nil {
		return nil, err
	}
	return push.Reconstitute(id, userID, endpoint, p256dh, auth, m.CreatedAt), nil
}
