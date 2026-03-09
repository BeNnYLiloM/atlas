-- Migration: 017_channel_categories
-- Description: Категории каналов воркспейса

CREATE TABLE IF NOT EXISTS channel_categories (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    name        VARCHAR(100) NOT NULL,
    position    INT NOT NULL DEFAULT 0,
    created_at  TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_channel_categories_workspace ON channel_categories(workspace_id);

ALTER TABLE channels
    ADD COLUMN IF NOT EXISTS category_id UUID REFERENCES channel_categories(id) ON DELETE SET NULL;
