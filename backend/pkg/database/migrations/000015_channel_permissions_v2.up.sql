-- Migration: 000015_channel_permissions_v2
-- Description: Упрощение прав доступа к каналу — только факт доступа без can_view/can_write.
--              Роли привязываются через workspace_roles.id вместо строки role.

DROP TABLE IF EXISTS channel_role_permissions;
DROP TABLE IF EXISTS channel_user_permissions;

CREATE TABLE channel_role_permissions (
    channel_id UUID NOT NULL REFERENCES channels(id) ON DELETE CASCADE,
    role_id    UUID NOT NULL REFERENCES workspace_roles(id) ON DELETE CASCADE,
    PRIMARY KEY (channel_id, role_id)
);

CREATE TABLE channel_user_permissions (
    channel_id UUID NOT NULL REFERENCES channels(id) ON DELETE CASCADE,
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    PRIMARY KEY (channel_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_channel_role_perms ON channel_role_permissions(channel_id);
CREATE INDEX IF NOT EXISTS idx_channel_user_perms ON channel_user_permissions(channel_id);
