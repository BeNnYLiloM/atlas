-- Добавляем поля для presence (статус онлайн/офлайн)
ALTER TABLE users 
ADD COLUMN IF NOT EXISTS status VARCHAR(20) DEFAULT 'offline',
ADD COLUMN IF NOT EXISTS last_seen TIMESTAMP WITH TIME ZONE;

CREATE INDEX IF NOT EXISTS idx_users_status ON users(status);

-- Функция для автоматического обновления last_seen при смене статуса
COMMENT ON COLUMN users.status IS 'Статус пользователя: online, away, offline';
COMMENT ON COLUMN users.last_seen IS 'Время последней активности пользователя';
