begin;

    CREATE TABLE user_session_preferences (
        user_id UUID NOT NULL UNIQUE,
        prefered_tenant BIGINT
    );

commit;
