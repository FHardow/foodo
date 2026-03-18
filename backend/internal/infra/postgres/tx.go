package postgres

import (
	"context"

	"gorm.io/gorm"
)

type txKey struct{}

// TxManager wraps transactions, propagating them via context.
type TxManager struct{ db *gorm.DB }

func NewTxManager(db *gorm.DB) *TxManager {
	return &TxManager{db: db}
}

func (t *TxManager) WithTx(ctx context.Context, fn func(ctx context.Context) error) error {
	return t.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(context.WithValue(ctx, txKey{}, tx))
	})
}

// dbFromCtx returns the transaction-scoped DB if present, otherwise the base DB.
func dbFromCtx(ctx context.Context, base *gorm.DB) *gorm.DB {
	if tx, ok := ctx.Value(txKey{}).(*gorm.DB); ok {
		return tx
	}
	return base.WithContext(ctx)
}
