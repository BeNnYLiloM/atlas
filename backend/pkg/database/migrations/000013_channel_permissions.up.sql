-- Migration: 000013_channel_permissions
-- Description: Первичные права доступа к каналу по ролям и участникам.
--              Заменена в 000015 на упрощённую версию без can_view/can_write.

CREATE TABLE IF NOT EXISTS channel_role_permissions (
    channel_id  UUID REFERENCES channels(id) ON DELETE CASCADE,
    role        VARCHAR(20) NOT NULL,
    can_view    BOOLEAN NOT NULL DEFAULT TRUE,
    can_write   BOOLEAN NOT NULL DEFAULT TRUE,
    PRIMARY KEY (channel_id, role)
);

CREATE TABLE IF NOT EXISTS channel_user_permissions (
    channel_id  UUID REFERENCES channels(id) ON DELETE CASCADE,
    user_id     UUID REFERENCES users(id) ON DELETE CASCADE,
    can_view    BOOLEAN NOT NULL DEFAULT TRUE,
    can_write   BOOLEAN NOT NULL DEFAULT TRUE,
    PRIMARY KEY (channel_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_channel_role_perms ON channel_role_permissions(channel_id);
CREATE INDEX IF NOT EXISTS idx_channel_user_perms ON channel_user_permissions(channel_id);
