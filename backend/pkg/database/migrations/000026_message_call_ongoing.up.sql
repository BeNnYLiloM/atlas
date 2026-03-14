-- Migration: 000026_message_call_ongoing
-- Добавляем статус 'ongoing' для активного звонка

ALTER TABLE messages
    DROP CONSTRAINT IF EXISTS messages_call_fields_check;

ALTER TABLE messages
    DROP CONSTRAINT IF EXISTS messages_call_status_check;

-- Пересоздаём constraint с новым значением
ALTER TABLE messages
    ADD CONSTRAINT messages_call_status_check
    CHECK (call_status IS NULL OR call_status IN ('missed', 'ongoing', 'ended'));

ALTER TABLE messages
    ADD CONSTRAINT messages_call_fields_check
    CHECK (type != 'call' OR call_status IS NOT NULL);
