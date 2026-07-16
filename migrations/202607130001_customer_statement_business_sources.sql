ALTER TABLE customer_statement_items ADD COLUMN source_type varchar(32);
ALTER TABLE customer_statement_items ADD COLUMN source_id integer;
ALTER TABLE customer_statement_items ADD COLUMN source_no varchar(64);
ALTER TABLE customer_statement_items ADD COLUMN receivable_id integer;

UPDATE customer_statement_items
SET source_type = 'sales',
    source_id = sale_id,
    source_no = sale_no
WHERE (source_type IS NULL OR source_type = '') AND sale_id > 0;

DROP INDEX IF EXISTS ux_customer_statement_items_sale_open;
CREATE INDEX IF NOT EXISTS idx_customer_statement_items_receivable ON customer_statement_items(tenant_id, receivable_id);
CREATE UNIQUE INDEX IF NOT EXISTS ux_customer_statement_items_source_open
ON customer_statement_items(tenant_id, source_type, source_id)
WHERE deleted_at IS NULL AND source_type IS NOT NULL AND source_id IS NOT NULL;
