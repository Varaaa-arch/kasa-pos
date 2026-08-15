CREATE TABLE transactions (
    id UUID PRIMARY KEY,
    invoice_number TEXT NOT NULL UNIQUE,

    subtotal BIGINT NOT NULL CHECK (subtotal >= 0),
    discount BIGINT NOT NULL DEFAULT 0 CHECK (discount >= 0),
    tax BIGINT NOT NULL DEFAULT 0 CHECK (tax >= 0),
    service_charge BIGINT NOT NULL DEFAULT 0 CHECK (service_charge >= 0),
    total BIGINT NOT NULL CHECK (total >= 0),

    paid_amount BIGINT NOT NULL CHECK (paid_amount >= 0),
    change BIGINT NOT NULL CHECK (change >= 0),

    payment_method TEXT NOT NULL,
    status TEXT NOT NULL,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_transactions_created_at
    ON transactions (created_at DESC);

CREATE INDEX idx_transactions_status
    ON transactions (status);


CREATE TABLE transaction_items (
    id UUID PRIMARY KEY,
    transaction_id UUID NOT NULL
        REFERENCES transactions(id)
        ON DELETE CASCADE,

    product_id UUID NOT NULL
        REFERENCES products(id)
        ON DELETE RESTRICT,

    sku TEXT NOT NULL,
    name TEXT NOT NULL,

    quantity INTEGER NOT NULL
        CHECK (quantity > 0),

    unit_price BIGINT NOT NULL
        CHECK (unit_price >= 0),

    subtotal BIGINT NOT NULL
        CHECK (subtotal >= 0)
);

CREATE INDEX idx_transaction_items_transaction_id
    ON transaction_items (transaction_id);

CREATE INDEX idx_transaction_items_product_id
    ON transaction_items (product_id);
