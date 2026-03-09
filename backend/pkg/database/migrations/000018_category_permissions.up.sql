-- Migration: 018_category_permissions
-- Description: Приватность категорий + права доступа

ALTER TABLE channel_categories
    ADD COLUMN IF NOT EXISTS is_private BOOLEAN NOT NULL DEFAULT FALSE;

-- Роли с доступом к приватной категории
CREATE TABLE IF NOT EXISTS category_role_permissions (
    category_id UUID NOT NULL REFERENCES channel_categories(id) ON DELETE CASCADE,
    role_id     UUID NOT NULL REFERENCES workspace_roles(id) ON DELETE CASCADE,
    PRIMARY KEY (category_id, role_id)
);

-- Участники с индивидуальным доступом к приватной категории
CREATE TABLE IF NOT EXISTS category_user_permissions (
    category_id UUID NOT NULL REFERENCES channel_categories(id) ON DELETE CASCADE,
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    PRIMARY KEY (category_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_cat_role_perms_cat ON category_role_permissions(category_id);
CREATE INDEX IF NOT EXISTS idx_cat_user_perms_cat ON category_user_permissions(category_id);
