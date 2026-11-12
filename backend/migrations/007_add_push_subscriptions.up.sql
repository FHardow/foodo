CREATE TABLE push_subscriptions (
    id            UUID PRIMARY KEY,
    user_id       UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    endpoint      TEXT NOT NULL,        -- AES-256-GCM ciphertext, base64
    endpoint_hash TEXT NOT NULL UNIQUE, -- SHA-256(endpoint), hex — lookup/upsert key
    p256dh        TEXT NOT NULL,        -- ciphertext, base64
    auth          TEXT NOT NULL,        -- ciphertext, base64
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_push_subscriptions_user_id ON push_subscriptions(user_id);
