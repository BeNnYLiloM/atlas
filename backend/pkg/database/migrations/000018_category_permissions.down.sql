DROP TABLE IF EXISTS category_user_permissions;
DROP TABLE IF EXISTS category_role_permissions;
ALTER TABLE channel_categories DROP COLUMN IF EXISTS is_private;
