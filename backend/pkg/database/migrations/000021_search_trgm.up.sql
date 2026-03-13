-- pg_trgm даёт GIN-индекс для ILIKE '%..%' поиска по подстроке.
-- Без этого расширения ILIKE делает seq scan по всей таблице messages.
CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE INDEX IF NOT EXISTS idx_messages_content_trgm
    ON messages USING GIN(content gin_trgm_ops);
