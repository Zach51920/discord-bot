-- Disable all triggers
SET session_replication_role = 'replica';

DO
$$
    BEGIN
        DROP TABLE IF EXISTS guilds CASCADE;
        DROP TABLE IF EXISTS channels CASCADE;

        DROP TYPE IF EXISTS code_execution_mode;
        DROP TRIGGER IF EXISTS update_guilds_updated_at_trigger ON guilds;
        DROP TRIGGER IF EXISTS update_channels_updated_at_trigger ON channels;

        DROP FUNCTION IF EXISTS update_updated_at_column;
        DROP USER IF EXISTS overlord_bot;

        DROP OWNED BY discord_bot;
        DROP ROLE IF EXISTS discord_bot;

    EXCEPTION
        WHEN others THEN
            ROLLBACK;
            RAISE EXCEPTION 'transaction failed: %', SQLERRM USING ERRCODE = SQLSTATE;
    END
$$;

COMMIT;

-- Enable all triggers
SET session_replication_role = 'origin';
