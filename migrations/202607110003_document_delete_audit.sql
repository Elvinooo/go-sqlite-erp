-- Unified document delete audit and reverse-processing records.
ALTER TABLE purchase_orders ADD COLUMN deleted_by integer;
ALTER TABLE purchase_orders ADD COLUMN delete_reason varchar(500);

CREATE TABLE IF NOT EXISTS document_delete_records (
    id integer PRIMARY KEY AUTOINCREMENT,
    tenant_id integer NOT NULL DEFAULT 1,
    remark text,
    created_by integer,
    updated_by integer,
    created_at datetime NOT NULL,
    updated_at datetime NOT NULL,
    deleted_at datetime,
    document_type varchar(32) NOT NULL,
    document_id integer NOT NULL,
    document_no varchar(64) NOT NULL,
    delete_reason varchar(500) NOT NULL,
    delete_user_id integer NOT NULL,
    delete_user_name varchar(64),
    delete_time datetime NOT NULL,
    delete_status varchar(32) NOT NULL DEFAULT 'WAITING',
    before_data text NOT NULL,
    stock_processed integer NOT NULL DEFAULT 0,
    finance_processed integer NOT NULL DEFAULT 0,
    ip_address varchar(64),
    completed_at datetime,
    failed_reason varchar(500)
);

CREATE TABLE IF NOT EXISTS document_delete_details (
    id integer PRIMARY KEY AUTOINCREMENT,
    tenant_id integer NOT NULL DEFAULT 1,
    remark text,
    created_by integer,
    updated_by integer,
    created_at datetime NOT NULL,
    updated_at datetime NOT NULL,
    deleted_at datetime,
    record_id integer NOT NULL,
    detail_type varchar(32) NOT NULL,
    sku_id integer,
    sku_code varchar(64),
    sku_name varchar(128),
    warehouse_id integer,
    warehouse varchar(64),
    quantity decimal(18,4) NOT NULL DEFAULT 0,
    stock_change decimal(18,4) NOT NULL DEFAULT 0,
    finance_type varchar(32),
    finance_no varchar(64),
    amount decimal(18,4) NOT NULL DEFAULT 0,
    business_type varchar(32),
    business_id integer,
    business_no varchar(64)
);

CREATE UNIQUE INDEX IF NOT EXISTS ux_document_delete_records_doc
    ON document_delete_records(tenant_id, document_type, document_id);
CREATE INDEX IF NOT EXISTS idx_document_delete_records_no
    ON document_delete_records(tenant_id, document_no);
CREATE INDEX IF NOT EXISTS idx_document_delete_records_time
    ON document_delete_records(tenant_id, delete_time);
CREATE INDEX IF NOT EXISTS idx_document_delete_records_user
    ON document_delete_records(tenant_id, delete_user_id);
CREATE INDEX IF NOT EXISTS idx_document_delete_records_status
    ON document_delete_records(tenant_id, delete_status);
CREATE INDEX IF NOT EXISTS idx_document_delete_details_record
    ON document_delete_details(tenant_id, record_id, detail_type);

INSERT INTO permissions (tenant_id, code, name, module, type, method, path, status, created_at, updated_at)
SELECT 1, 'document_delete', '单据删除', 'document', 'api', 'DELETE', '/api/v1/:module/:id', 'active', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE tenant_id = 1 AND code = 'document_delete');

INSERT INTO menus (tenant_id, name, title, path, component, icon, type, permission_code, sort, visible, status, created_at, updated_at)
SELECT 1, 'document-delete-records', '单据删除记录', '/document-delete-records', 'mobile/FeaturePage', 'Document', 'menu', 'system.audit.view', 165, 1, 'active', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
WHERE NOT EXISTS (SELECT 1 FROM menus WHERE tenant_id = 1 AND name = 'document-delete-records');
