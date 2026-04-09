CREATE TABLE IF NOT EXISTS product_views (
    id         BIGSERIAL PRIMARY KEY,
    product_id BIGINT    NOT NULL,
    viewed_at  TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_product_views_product_id ON product_views (product_id);
CREATE INDEX IF NOT EXISTS idx_product_views_viewed_at  ON product_views (viewed_at);
