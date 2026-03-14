-- Migration: 000022_projects
-- Description: Система проектов — изолированные пространства внутри воркспейса

-- 1. Таблица проектов
CREATE TABLE IF NOT EXISTS projects (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    name         VARCHAR(100) NOT NULL,
    description  TEXT,
    icon_url     TEXT,
    is_archived  BOOLEAN NOT NULL DEFAULT FALSE,
    created_at   TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_projects_workspace ON projects(workspace_id);

-- 2. Участники проекта
-- Роль определяется через workspace_member_roles воркспейса.
-- is_lead — руководитель проекта, назначается ws owner или участником с ManageMembers.
CREATE TABLE IF NOT EXISTS project_members (
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    is_lead    BOOLEAN NOT NULL DEFAULT FALSE,
    joined_at  TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    PRIMARY KEY (project_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_project_members_user ON project_members(user_id);

-- 3. Привязка каналов и категорий к проекту
-- project_id IS NULL  → канал/категория принадлежит воркспейсу
-- project_id IS NOT NULL → принадлежит проекту (видим только project_members)
ALTER TABLE channels
    ADD COLUMN IF NOT EXISTS project_id UUID REFERENCES projects(id) ON DELETE CASCADE;

ALTER TABLE channel_categories
    ADD COLUMN IF NOT EXISTS project_id UUID REFERENCES projects(id) ON DELETE CASCADE;

CREATE INDEX IF NOT EXISTS idx_channels_project ON channels(project_id) WHERE project_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_categories_project ON channel_categories(project_id) WHERE project_id IS NOT NULL;

-- 4. Правка триггера: при вступлении в воркспейс добавлять
--    только в каналы воркспейса (project_id IS NULL)
CREATE OR REPLACE FUNCTION add_member_to_public_channels()
RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO channel_members (user_id, channel_id, joined_at)
    SELECT NEW.user_id, c.id, NOW()
    FROM channels c
    WHERE c.workspace_id = NEW.workspace_id
      AND c.is_private = FALSE
      AND c.project_id IS NULL
    ON CONFLICT (user_id, channel_id) DO NOTHING;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- 5. Правка триггера: при создании публичного воркспейс-канала
--    добавлять только участников воркспейса (не для каналов проекта)
CREATE OR REPLACE FUNCTION add_workspace_members_to_channel()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.is_private = FALSE AND NEW.project_id IS NULL THEN
        INSERT INTO channel_members (user_id, channel_id, joined_at)
        SELECT wm.user_id, NEW.id, NOW()
        FROM workspace_members wm
        WHERE wm.workspace_id = NEW.workspace_id
        ON CONFLICT (user_id, channel_id) DO NOTHING;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
