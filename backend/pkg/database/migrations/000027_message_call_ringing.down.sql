ALTER TABLE messages
    DROP CONSTRAINT IF EXISTS messages_call_fields_check,
    DROP CONSTRAINT IF EXISTS messages_call_status_check;

ALTER TABLE messages
    ADD CONSTRAINT messages_call_status_check
    CHECK (call_status IS NULL OR call_status IN ('missed', 'ongoing', 'ended'));

ALTER TABLE messages
    ADD CONSTRAINT messages_call_fields_check
    CHECK (type != 'call' OR call_status IS NOT NULL);
