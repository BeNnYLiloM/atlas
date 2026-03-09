-- Migration: 000014_workspace_roles
-- Description: Кастомные роли воркспейса (аналог Discord)

CREATE TABLE IF NOT EXISTS workspace_roles (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    name         VARCHAR(100) NOT NULL,
    color        VARCHAR(7)  NOT NULL DEFAULT '#99aab5', -- hex цвет
    position     INT         NOT NULL DEFAULT 0,         -- чем выше — тем приоритетнее
    is_system    BOOLEAN     NOT NULL DEFAULT FALSE,     -- owner/admin/@everyone — нельзя удалить
    -- Permissions (JSON-объект с булевыми флагами)
    permissions  JSONB       NOT NULL DEFAULT '{}',
    created_at   TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE (workspace_id, name)
);

-- Назначение кастомных ролей участникам (многие-ко-многим)
CREATE TABLE IF NOT EXISTS workspace_member_roles (
    workspace_id UUID NOT NULL,
    user_id      UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id      UUID NOT NULL REFERENCES workspace_roles(id) ON DELETE CASCADE,
    PRIMARY KEY (workspace_id, user_id, role_id)
);

CREATE INDEX IF NOT EXISTS idx_workspace_roles_ws ON workspace_roles(workspace_id, position DESC);
CREATE INDEX IF NOT EXISTS idx_member_roles_user ON workspace_member_roles(workspace_id, user_id);
