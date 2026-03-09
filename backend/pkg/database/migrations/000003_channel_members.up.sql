-- Таблица для отслеживания прочитанных сообщений в каналах
CREATE TABLE IF NOT EXISTS channel_members (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    channel_id UUID NOT NULL REFERENCES channels(id) ON DELETE CASCADE,
    last_read_message_id UUID REFERENCES messages(id) ON DELETE SET NULL,
    last_read_at TIMESTAMP WITH TIME ZONE,
    joined_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    PRIMARY KEY (user_id, channel_id)
);

CREATE INDEX idx_channel_members_user ON channel_members(user_id);
CREATE INDEX idx_channel_members_channel ON channel_members(channel_id);
CREATE INDEX idx_channel_members_last_read ON channel_members(last_read_message_id);

-- При добавлении участника в workspace, автоматически добавляем его во все публичные каналы
CREATE OR REPLACE FUNCTION add_member_to_public_channels()
RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO channel_members (user_id, channel_id, joined_at)
    SELECT NEW.user_id, c.id, NOW()
    FROM channels c
    WHERE c.workspace_id = NEW.workspace_id 
      AND c.is_private = false
    ON CONFLICT (user_id, channel_id) DO NOTHING;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_add_member_to_channels
AFTER INSERT ON workspace_members
FOR EACH ROW
EXECUTE FUNCTION add_member_to_public_channels();

-- При создании нового канала, добавляем всех участников workspace
CREATE OR REPLACE FUNCTION add_workspace_members_to_channel()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.is_private = false THEN
        INSERT INTO channel_members (user_id, channel_id, joined_at)
        SELECT wm.user_id, NEW.id, NOW()
        FROM workspace_members wm
        WHERE wm.workspace_id = NEW.workspace_id
        ON CONFLICT (user_id, channel_id) DO NOTHING;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_add_members_to_new_channel
AFTER INSERT ON channels
FOR EACH ROW
EXECUTE FUNCTION add_workspace_members_to_channel();
