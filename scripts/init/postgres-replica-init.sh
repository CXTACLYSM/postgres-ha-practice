#!/bin/bash
# =============================================================================
# PostgreSQL REPLICA - INITIALIZATION SCRIPT
# =============================================================================
# This script runs as the entrypoint for each replica container.
# It is intentionally generic — all instance-specific parameters come
# from environment variables set in compose.ha.yml:
#
#   PRIMARY_HOST       — hostname of the primary (postgres_primary)
#   REPLICATOR_USER    — replication user created by primary init script
#   REPLICATOR_PASSWORD
#   REPLICA_SLOT_NAME  — physical replication slot name, unique per replica:
#                        replica_1 → replica_slot_1
#                        replica_2 → replica_slot_2
#
# This single script serves both replica_1 and replica_2 — Docker Compose
# runs it twice with different environment variables. That is the "loop":
# parameterization through the environment, not through bash iteration.
# =============================================================================

set -e

echo "=== Replica init: slot=${REPLICA_SLOT_NAME} primary=${PRIMARY_HOST} ==="

# Check if data directory is already initialized.
# PG_VERSION file exists only after a successful pg_basebackup or initdb.
# On container restart (not first start) we skip straight to PostgreSQL startup.
if [ -s /var/lib/postgresql/data/PG_VERSION ]; then
    echo "Replica: Data directory already initialized, starting PostgreSQL..."
else
    echo "Replica: Initializing from primary via pg_basebackup..."

    # Wait for primary to accept replication connections.
    # pg_isready checks the TCP port only — it does not authenticate.
    # We loop until primary is ready before attempting pg_basebackup.
    until PGPASSWORD="${REPLICATOR_PASSWORD}" pg_isready \
        -h "${PRIMARY_HOST}" \
        -U "${REPLICATOR_USER}"; do
        echo "Waiting for primary at ${PRIMARY_HOST}..."
        sleep 2
    done

    echo "Primary is ready. Starting base backup with slot: ${REPLICA_SLOT_NAME}..."

    # pg_basebackup flags:
    #   -h  primary host
    #   -U  replication user (must have REPLICATION attribute)
    #   -D  destination data directory
    #   -Fp plain format (one file per relation, same layout as primary)
    #   -Xs stream WAL during backup so replica can start immediately after
    #   -P  show progress
    #   -R  write standby.signal + primary_conninfo into postgresql.auto.conf
    #       this is what tells PostgreSQL "you are a standby, connect here"
    #   -S  bind to this replication slot on primary
    #       the slot prevents primary from discarding WAL until this replica
    #       confirms receipt — critical for replicas that may lag behind
    #
    # REPLICA_SLOT_NAME comes from compose environment block.
    # replica_1 receives replica_slot_1, replica_2 receives replica_slot_2.
    # Two independent slots = primary tracks each replica's progress separately.
    PGPASSWORD="${REPLICATOR_PASSWORD}" pg_basebackup \
        -h "${PRIMARY_HOST}" \
        -U "${REPLICATOR_USER}" \
        -D /var/lib/postgresql/data \
        -Fp -Xs -P -R \
        -S "${REPLICA_SLOT_NAME}"

    # Copy our custom configs into the data directory.
    # postgresql.auto.conf written by -R above has higher priority than
    # postgresql.conf, so primary_conninfo and primary_slot_name set by
    # pg_basebackup will NOT be overwritten by our postgresql.conf copy.
    # Our conf only adds hot_standby settings, logging, etc.
    cp /etc/postgresql/postgresql.conf /var/lib/postgresql/data/
    cp /etc/postgresql/pg_hba.conf /var/lib/postgresql/data/

    # pg_basebackup runs as root (container user: root).
    # PostgreSQL refuses to start if data directory is not owned by postgres.
    chown -R postgres:postgres /var/lib/postgresql/data
    chmod 700 /var/lib/postgresql/data

    echo "Replica: Base backup completed! Slot: ${REPLICA_SLOT_NAME}"
fi

# su-exec switches from root to postgres before exec.
# exec replaces the shell process — PostgreSQL becomes PID 1 in the container,
# which means Docker signals (SIGTERM on stop) go directly to PostgreSQL.
exec su-exec postgres postgres \
    -c config_file=/var/lib/postgresql/data/postgresql.conf \
    -c hba_file=/var/lib/postgresql/data/pg_hba.conf