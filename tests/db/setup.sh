#!/bin/bash
set -e

# Load environment variables from .env.test if present
# Determine script directory to locate env file robustly
SCRIPT_DIR=$(dirname "$(readlink -f "$0")")
ENV_FILE="$SCRIPT_DIR/../env/.env.test"

if [ -f "$ENV_FILE" ]; then
    export $(grep -v '^#' "$ENV_FILE" | xargs)
fi

# Fallback defaults if not set in env file
DB_URL=${DB_URL:-"postgres://postgres:postgres@localhost:5432/templatedb"}

echo "Initializing Test Database at $DB_URL..."

# apply database initialization
psql "$DB_URL" -qf tests/db/init.sql

echo "Test Database Initialized Successfully."
