ALTER TABLE users
ADD COLUMN IF NOT EXISTS custom_status VARCHAR(100);

COMMENT ON COLUMN users.custom_status IS 'Кастомный текстовый статус пользователя';
