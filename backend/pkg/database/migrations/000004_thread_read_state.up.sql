-- Добавляем поле для отслеживания последнего прочитанного сообщения в треде
ALTER TABLE channel_members 
ADD COLUMN IF NOT EXISTS thread_last_reads JSONB DEFAULT '{}'::jsonb;

-- Индекс для быстрого доступа к thread reads
CREATE INDEX IF NOT EXISTS idx_channel_members_thread_reads ON channel_members USING GIN (thread_last_reads);

-- Комментарий для пояснения структуры
COMMENT ON COLUMN channel_members.thread_last_reads IS 'JSON объект формата: {"parent_message_id": "last_read_message_id"}';
