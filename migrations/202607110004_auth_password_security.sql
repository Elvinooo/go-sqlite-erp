ALTER TABLE users ADD COLUMN must_change_password integer NOT NULL DEFAULT 0;
ALTER TABLE users ADD COLUMN password_changed_at datetime;
ALTER TABLE users ADD COLUMN password_version integer NOT NULL DEFAULT 0;

UPDATE users
SET password_changed_at = COALESCE(updated_at, created_at, CURRENT_TIMESTAMP)
WHERE password_changed_at IS NULL;

UPDATE system_settings
SET setting_value = '8'
WHERE setting_key = 'password.min_length';
