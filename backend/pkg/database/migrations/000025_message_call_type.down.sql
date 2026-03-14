ALTER TABLE messages
    DROP CONSTRAINT IF EXISTS messages_call_fields_check,
    DROP COLUMN IF EXISTS call_duration_sec,
    DROP COLUMN IF EXISTS call_status,
    DROP COLUMN IF EXISTS type;
