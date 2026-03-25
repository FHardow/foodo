ALTER TABLE users
    ADD COLUMN role TEXT NOT NULL DEFAULT 'customer'
        CHECK (role IN ('customer', 'owner'));

-- Seed a known customer and a known owner for local development.
-- These UUIDs match the constants in the frontend.
INSERT INTO users (id, name, email, phone, role)
VALUES
    ('00000000-0000-0000-0000-000000000001', 'Alice Customer', 'alice@example.com', '', 'customer'),
    ('00000000-0000-0000-0000-000000000002', 'Bob Owner',      'bob@example.com',   '', 'owner')
ON CONFLICT (id) DO NOTHING;
