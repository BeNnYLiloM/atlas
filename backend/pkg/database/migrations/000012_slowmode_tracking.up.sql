-- Migration: 000012_slowmode_tracking
-- Description: Отслеживание времени последнего сообщения для slowmode

ALTER TABLE channel_members
    ADD COLUMN IF NOT EXISTS last_message_at TIMESTAMP WITH TIME ZONE;
