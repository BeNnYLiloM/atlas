-- Migration: 009_workspace_settings
-- Description: Настройки воркспейса: описание, иконка; никнейм участника

ALTER TABLE workspaces
    ADD COLUMN IF NOT EXISTS description TEXT,
    ADD COLUMN IF NOT EXISTS icon_url TEXT;

ALTER TABLE workspace_members
    ADD COLUMN IF NOT EXISTS nickname TEXT;
