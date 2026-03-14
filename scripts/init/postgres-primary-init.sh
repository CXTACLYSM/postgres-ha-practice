#!/bin/bash
# =============================================================================
# PostgreSQL PRIMARY - INITIALIZATION SCRIPT
# =============================================================================
# Executed ONCE when the container starts with empty volume
# Uses environment variables from docker-compose
# =============================================================================

set -e

echo "=== Initializing PostgreSQL PRIMARY ==="
echo "Database: $POSTGRES_DB"
echo "Write user: $APP_WRITE_USER"
echo "Read user: $APP_READ_USER"
echo "Replicator: $POSTGRES_REPLICATOR_USER"

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL

    -- =========================================================================
    -- EXTENSIONS
    -- =========================================================================
    CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
    CREATE EXTENSION IF NOT EXISTS "pgcrypto";

    -- =========================================================================
    -- APPLICATION USERS
    -- =========================================================================
    -- Write user: full data access (INSERT, UPDATE, DELETE)
    -- Used by Command part of CQRS
    DO \$\$
    BEGIN
        IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = '$APP_WRITE_USER') THEN
            CREATE ROLE $APP_WRITE_USER WITH LOGIN PASSWORD '$APP_WRITE_PASSWORD';
            RAISE NOTICE 'Created user: $APP_WRITE_USER';
        ELSE
            RAISE NOTICE 'User already exists: $APP_WRITE_USER';
        END IF;
    END
    \$\$;

    -- Read user: SELECT only
    -- Used by Query part of CQRS (connects to replica)
    DO \$\$
    BEGIN
        IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = '$APP_READ_USER') THEN
            CREATE ROLE $APP_READ_USER WITH LOGIN PASSWORD '$APP_READ_PASSWORD';
            RAISE NOTICE 'Created user: $APP_READ_USER';
        ELSE
            RAISE NOTICE 'User already exists: $APP_READ_USER';
        END IF;
    END
    \$\$;

    -- Grant permissions
    GRANT CONNECT ON DATABASE $POSTGRES_DB TO $APP_WRITE_USER, $APP_READ_USER;
    GRANT USAGE ON SCHEMA public TO $APP_WRITE_USER, $APP_READ_USER;
    GRANT CREATE ON SCHEMA public TO $APP_WRITE_USER;

    -- Write user: all data operations
    GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO $APP_WRITE_USER;
    ALTER DEFAULT PRIVILEGES FOR ROLE $APP_WRITE_USER IN SCHEMA public
        GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO $APP_WRITE_USER;
    GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO $APP_WRITE_USER;
    ALTER DEFAULT PRIVILEGES FOR ROLE $APP_WRITE_USER IN SCHEMA public
        GRANT USAGE, SELECT ON SEQUENCES TO $APP_WRITE_USER;

    -- Read user: read only
    GRANT SELECT ON ALL TABLES IN SCHEMA public TO $APP_READ_USER;
    ALTER DEFAULT PRIVILEGES FOR ROLE $APP_WRITE_USER IN SCHEMA public
        GRANT SELECT ON TABLES TO $APP_READ_USER;

    -- =========================================================================
    -- REPLICATION USER
    -- =========================================================================
    -- Special user ONLY for replication
    -- REPLICATION is a separate attribute, not related to table permissions
    DO \$\$
    BEGIN
        IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = '$POSTGRES_REPLICATOR_USER') THEN
            CREATE ROLE $POSTGRES_REPLICATOR_USER WITH REPLICATION LOGIN PASSWORD '$POSTGRES_REPLICATOR_PASSWORD';
            RAISE NOTICE 'Created replication user: $POSTGRES_REPLICATOR_USER';
        ELSE
            RAISE NOTICE 'Replication user already exists: $POSTGRES_REPLICATOR_USER';
        END IF;
    END
    \$\$;

    -- =========================================================================
    -- REPLICATION SLOT
    -- =========================================================================
    -- Slot guarantees that WAL segments won't be deleted until replica
    -- confirms receipt
    DO \$\$
    BEGIN
        IF NOT EXISTS (SELECT 1 FROM pg_replication_slots WHERE slot_name = 'replica_slot_1') THEN
            PERFORM pg_create_physical_replication_slot('replica_slot_1');
            RAISE NOTICE 'Created replication slot: replica_slot_1';
        ELSE
            RAISE NOTICE 'Replication slot already exists: replica_slot_1';
        END IF;
        IF NOT EXISTS (SELECT 1 FROM pg_replication_slots WHERE slot_name = 'replica_slot_2') THEN
                    PERFORM pg_create_physical_replication_slot('replica_slot_2');
              RAISE NOTICE 'Created replication slot: replica_slot_2';
        ELSE
              RAISE NOTICE 'Replication slot already exists: replica_slot_2';
        END IF;
    END
    \$\$;

    -- =========================================================================
    -- VERIFICATION
    -- =========================================================================
    \echo ''
    \echo '=== Created users ==='
    SELECT rolname, rolcanlogin, rolreplication
    FROM pg_roles
    WHERE rolname IN ('$APP_WRITE_USER', '$APP_READ_USER', '$POSTGRES_REPLICATOR_USER');

    \echo ''
    \echo '=== Replication slots ==='
    SELECT slot_name, slot_type, active FROM pg_replication_slots;

EOSQL

echo ""
echo "=== PostgreSQL PRIMARY initialized successfully ==="