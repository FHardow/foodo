UPDATE orders SET status = 'finished' WHERE status = 'archived';

ALTER TABLE orders DROP CONSTRAINT IF EXISTS orders_status_check;
ALTER TABLE orders ADD CONSTRAINT orders_status_check
    CHECK (status IN ('pending', 'created', 'accepted', 'ongoing', 'finished'));
