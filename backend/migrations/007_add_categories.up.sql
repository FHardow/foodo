CREATE TABLE categories (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name       TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Category names are unique regardless of case ("Fruit" == "fruit").
CREATE UNIQUE INDEX idx_categories_name_lower ON categories (lower(name));

CREATE TABLE product_categories (
    product_id  UUID NOT NULL REFERENCES products(id)   ON DELETE CASCADE,
    category_id UUID NOT NULL REFERENCES categories(id) ON DELETE CASCADE,
    PRIMARY KEY (product_id, category_id)
);

CREATE INDEX idx_product_categories_category_id ON product_categories(category_id);
