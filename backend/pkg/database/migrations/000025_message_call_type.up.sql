-- Migration: 000025_message_call_type
-- Description: Поддержка call-сообщений в истории чата

ALTER TABLE messages
    ADD COLUMN type            VARCHAR(10)  NOT NULL DEFAULT 'text' CHECK (type IN ('text', 'call')),
    ADD COLUMN call_status     VARCHAR(10)  CHECK (call_status IN ('missed', 'ended')),
    ADD COLUMN call_duration_sec INT;

-- call_status обязателен для call-сообщений
ALTER TABLE messages
    ADD CONSTRAINT messages_call_fields_check
    CHECK (type != 'call' OR call_status IS NOT NULL);
