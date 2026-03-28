-- Reverse data migration
UPDATE orders SET status = 'fulfilled' WHERE status = 'finished';
UPDATE orders SET status = 'confirmed' WHERE status = 'created';
UPDATE orders SET status = 'confirmed' WHERE status IN ('accepted', 'ongoing');

-- Restore original constraint
ALTER TABLE orders DROP CONSTRAINT IF EXISTS orders_status_check;
ALTER TABLE orders ADD CONSTRAINT orders_status_check
    CHECK (status IN ('pending', 'confirmed', 'fulfilled', 'cancelled'));
