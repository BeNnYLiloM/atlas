-- Migration: 000011_channel_settings
-- Description: Настройки канала: топик, slowmode, позиция; уведомления в channel_members

ALTER TABLE channels
    ADD COLUMN IF NOT EXISTS topic              TEXT,
    ADD COLUMN IF NOT EXISTS slowmode_seconds   INT     NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS position           INT     NOT NULL DEFAULT 0;

ALTER TABLE channel_members
    ADD COLUMN IF NOT EXISTS notification_level VARCHAR(20) NOT NULL DEFAULT 'all';
-- notification_level: 'all' | 'mentions' | 'nothing'
