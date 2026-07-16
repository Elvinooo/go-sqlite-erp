-- One-time SQLite migration for databases created before this feature.
-- The application AutoMigrate path applies the same schema changes automatically.
ALTER TABLE sales_orders ADD COLUMN deleted_by integer;
ALTER TABLE sales_orders ADD COLUMN delete_reason varchar(500) NOT NULL DEFAULT '';
ALTER TABLE inventory_movements ADD COLUMN warehouse varchar(64) NOT NULL DEFAULT '主仓库';

CREATE TABLE IF NOT EXISTS sales_order_deletions (
    id integer PRIMARY KEY AUTOINCREMENT,
    tenant_id integer NOT NULL DEFAULT 1,
    original_order_id integer NOT NULL,
    order_no varchar(64) NOT NULL,
    customer_id integer,
    customer_name varchar(128),
    items_json text NOT NULL,
    deleted_by integer NOT NULL,
    deleted_at datetime NOT NULL,
    delete_reason varchar(500) NOT NULL,
    snapshot_json text NOT NULL,
    created_at datetime NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS ux_sales_order_deletions_order
    ON sales_order_deletions(tenant_id, original_order_id);
CREATE INDEX IF NOT EXISTS idx_sales_order_deletions_time
    ON sales_order_deletions(tenant_id, deleted_at);
CREATE INDEX IF NOT EXISTS idx_sales_orders_deleted_by
    ON sales_orders(deleted_by);
