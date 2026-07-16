CREATE TABLE IF NOT EXISTS customer_statements (
    id integer PRIMARY KEY AUTOINCREMENT,
    tenant_id integer NOT NULL DEFAULT 1,
    statement_no varchar(64) NOT NULL,
    customer_id integer NOT NULL,
    customer_name varchar(128) NOT NULL,
    contact_name varchar(64),
    contact_phone varchar(32),
    start_date datetime NOT NULL,
    end_date datetime NOT NULL,
    total_amount decimal(18,4) NOT NULL DEFAULT 0,
    received_amount decimal(18,4) NOT NULL DEFAULT 0,
    unpaid_amount decimal(18,4) NOT NULL DEFAULT 0,
    cumulative_debt decimal(18,4) NOT NULL DEFAULT 0,
    status varchar(32) NOT NULL DEFAULT 'unconfirmed',
    confirmed_at datetime,
    settled_at datetime,
    remark text,
    created_by integer,
    updated_by integer,
    created_at datetime,
    updated_at datetime,
    deleted_at datetime
);

CREATE TABLE IF NOT EXISTS customer_statement_items (
    id integer PRIMARY KEY AUTOINCREMENT,
    tenant_id integer NOT NULL DEFAULT 1,
    statement_id integer NOT NULL,
    sale_id integer NOT NULL,
    sale_no varchar(64) NOT NULL,
    sale_date datetime NOT NULL,
    source_type varchar(32),
    source_id integer,
    source_no varchar(64),
    receivable_id integer,
    product_name varchar(128),
    quantity decimal(18,4) NOT NULL DEFAULT 0,
    total_amount decimal(18,4) NOT NULL DEFAULT 0,
    received_amount decimal(18,4) NOT NULL DEFAULT 0,
    unpaid_amount decimal(18,4) NOT NULL DEFAULT 0,
    settlement_status varchar(32),
    remark text,
    created_by integer,
    updated_by integer,
    created_at datetime,
    updated_at datetime,
    deleted_at datetime
);

CREATE UNIQUE INDEX IF NOT EXISTS ux_customer_statements_tenant_no ON customer_statements(tenant_id, statement_no);
CREATE INDEX IF NOT EXISTS idx_customer_statements_customer_period ON customer_statements(tenant_id, customer_id, start_date, end_date, status);
CREATE INDEX IF NOT EXISTS idx_customer_statement_items_statement ON customer_statement_items(tenant_id, statement_id);
CREATE INDEX IF NOT EXISTS idx_customer_statement_items_receivable ON customer_statement_items(tenant_id, receivable_id);
CREATE UNIQUE INDEX IF NOT EXISTS ux_customer_statement_items_source_open ON customer_statement_items(tenant_id, source_type, source_id) WHERE deleted_at IS NULL AND source_type IS NOT NULL AND source_id IS NOT NULL;
