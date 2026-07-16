ALTER TABLE sales_order_items ADD COLUMN inventory_batch_id integer;
ALTER TABLE sales_order_items ADD COLUMN purchase_order_item_id integer;

UPDATE sales_order_items
SET inventory_batch_id = (
    SELECT sca.inventory_batch_id
    FROM sales_cost_allocations sca
    WHERE sca.tenant_id = sales_order_items.tenant_id
      AND sca.sales_order_item_id = sales_order_items.id
      AND sca.deleted_at IS NULL
    ORDER BY sca.id
    LIMIT 1
)
WHERE inventory_batch_id IS NULL;

UPDATE sales_order_items
SET purchase_order_item_id = (
    SELECT ib.purchase_order_item_id
    FROM inventory_batches ib
    WHERE ib.tenant_id = sales_order_items.tenant_id
      AND ib.id = sales_order_items.inventory_batch_id
      AND ib.deleted_at IS NULL
    LIMIT 1
)
WHERE purchase_order_item_id IS NULL AND inventory_batch_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_sales_items_purchase ON sales_order_items(tenant_id, purchase_order_id, purchase_order_item_id, inventory_batch_id);
