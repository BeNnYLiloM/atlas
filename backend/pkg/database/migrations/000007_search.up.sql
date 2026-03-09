-- Добавляем tsvector колонку для полнотекстового поиска
ALTER TABLE messages
ADD COLUMN IF NOT EXISTS search_vector tsvector;

-- GIN индекс для быстрого поиска
CREATE INDEX IF NOT EXISTS idx_messages_search ON messages USING GIN(search_vector);

-- Заполняем существующие сообщения
UPDATE messages SET search_vector = to_tsvector('russian', COALESCE(content, ''));

-- Триггер для автоматического обновления при insert/update
CREATE OR REPLACE FUNCTION messages_search_trigger() RETURNS trigger AS $$
BEGIN
  NEW.search_vector := to_tsvector('russian', COALESCE(NEW.content, ''));
  RETURN NEW;
END
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS messages_search_update ON messages;
CREATE TRIGGER messages_search_update
BEFORE INSERT OR UPDATE OF content ON messages
FOR EACH ROW EXECUTE FUNCTION messages_search_trigger();
