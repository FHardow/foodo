-- Migrate existing data to new status values
UPDATE orders SET status = 'created'  WHERE status = 'confirmed';
UPDATE orders SET status = 'finished' WHERE status IN ('fulfilled', 'cancelled');

-- Replace the status check constraint with new values
ALTER TABLE orders DROP CONSTRAINT IF EXISTS orders_status_check;
ALTER TABLE orders ADD CONSTRAINT orders_status_check
    CHECK (status IN ('pending', 'created', 'accepted', 'ongoing', 'finished'));
