#!/bin/bash
source .env

export MIGRATION_DSN="host=pg-auth port=5432 dbname=$PG_DATABASE_NAME user=$PG_USER password=$PG_PASSWORD sslmode=disable"

sleep 5 && goose -dir "${MIGRATION_DIR}" postgres "${MIGRATION_DSN}" up -v