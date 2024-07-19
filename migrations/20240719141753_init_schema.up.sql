-- Disable all triggers
SET session_replication_role = 'replica';

-- Start a transaction
BEGIN;

--
-- Create custom types if they don't already exist
--

DO
$$
    BEGIN
        CREATE TYPE code_execution_mode AS ENUM ('DISABLED', 'AUTO','MANUAL');
    EXCEPTION
        WHEN duplicate_object THEN RAISE NOTICE '%, skipping', SQLERRM USING ERRCODE = SQLSTATE;
    END
$$;

--
-- Define database schema
--

CREATE TABLE IF NOT EXISTS guilds
(
    guild_id   TEXT      NOT NULL,
    joined_at  TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (guild_id)
);

CREATE TABLE IF NOT EXISTS channels
(
    channel_id TEXT                NOT NULL,
    guild_id   TEXT                NOT NULL,
    updated_at TIMESTAMP           NOT NULL DEFAULT NOW(),
    code_exec  code_execution_mode NOT NULL DEFAULT 'DISABLED',
    PRIMARY KEY (channel_id),
    CONSTRAINT fk_guild
        FOREIGN KEY (guild_id)
            REFERENCES guilds (guild_id)
            ON DELETE CASCADE
);

COMMENT ON COLUMN channels.code_exec is 'Indicates how code execution should be managed for the channel';

--
-- Create functions and triggers
--

-- Create function to set the updated_at column to the current time
CREATE OR REPLACE FUNCTION update_updated_at_column()
    RETURNS TRIGGER AS
$$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create triggers to call the update_updated_at_column function when a table is updated
CREATE OR REPLACE TRIGGER update_guilds_updated_at_trigger
    BEFORE UPDATE
    ON guilds
    FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

CREATE OR REPLACE TRIGGER update_channels_updated_at_trigger
    BEFORE UPDATE
    ON channels
    FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

--
-- Create users and roles if they don't already exist
--

DO
$$
    BEGIN
        CREATE ROLE discord_bot;
    EXCEPTION
        WHEN duplicate_object THEN RAISE NOTICE '%, skipping', SQLERRM USING ERRCODE = SQLSTATE;
    END
$$;

GRANT USAGE ON SCHEMA public TO discord_bot;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO discord_bot;

GRANT SELECT, INSERT, UPDATE, DELETE ON guilds TO discord_bot;
GRANT SELECT, INSERT, UPDATE, DELETE ON channels TO discord_bot;

DO
$$
    BEGIN
        CREATE USER overlord_bot;
    EXCEPTION
        WHEN duplicate_object THEN RAISE NOTICE '%, skipping', SQLERRM USING ERRCODE = SQLSTATE;
    END
$$;

GRANT discord_bot TO overlord_bot;

ALTER USER overlord_bot WITH PASSWORD 'password';

-- Commit transaction
COMMIT;

-- Enable all triggers
SET session_replication_role = 'origin';