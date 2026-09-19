CREATE TABLE stock_items (
    sku TEXT PRIMARY KEY,
    available INTEGER NOT NULL CHECK (available >= 0),
    reserved INTEGER NOT NULL DEFAULT 0 CHECK (reserved >= 0),
    minimum_quantity INTEGER NOT NULL DEFAULT 0 CHECK (minimum_quantity >= 0),
    updated_at TIMESTAMPTZ NOT NULL
);
