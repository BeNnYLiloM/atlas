DROP TABLE IF EXISTS dm_channels;

ALTER TABLE channels
    DROP CONSTRAINT IF EXISTS channels_type_check;
