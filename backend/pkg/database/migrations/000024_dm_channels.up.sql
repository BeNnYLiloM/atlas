-- Migration: 000024_dm_channels
-- Description: Добавление личных сообщений (DM)

-- 1. Разрешаем тип 'dm' для каналов
ALTER TABLE channels
    ADD CONSTRAINT channels_type_check CHECK (type IN ('text', 'voice', 'dm'));

-- 2. Таблица для быстрого lookup DM-пары
-- Инвариант: user1_id < user2_id — гарантирует уникальность без дублей
-- и исключает race condition при параллельном создании.
CREATE TABLE dm_channels (
    channel_id UUID PRIMARY KEY REFERENCES channels(id) ON DELETE CASCADE,
    user1_id   UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    user2_id   UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    CONSTRAINT dm_users_ordered  CHECK (user1_id < user2_id),
    CONSTRAINT dm_users_unique   UNIQUE (user1_id, user2_id)
);

-- Поиск всех DM пользователя (с любой стороны)
CREATE INDEX idx_dm_channels_user1 ON dm_channels(user1_id);
CREATE INDEX idx_dm_channels_user2 ON dm_channels(user2_id);
