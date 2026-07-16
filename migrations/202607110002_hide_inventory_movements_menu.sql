-- Hide standalone inventory movement menu; movement history remains available in inventory detail.
UPDATE menus
SET visible = 0
WHERE name = 'inventory-movements'
   OR path = '/inventory-movements';
