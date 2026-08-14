CREATE TABLE products (
    id UUID PRIMARY KEY,
    sku TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    price BIGINT NOT NULL,
    stock INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT products_price_non_negative
        CHECK (price >= 0),

    CONSTRAINT products_stock_non_negative
        CHECK (stock >= 0)
);

CREATE INDEX idx_products_name
    ON products (name);

CREATE INDEX idx_products_sku
    ON products (sku);