#!/bin/sh
set -e

mkdir -p data

DB_FILE="data/goodbuzz.db"

if [ ! -f "$DB_FILE" ]; then
    echo "Initializing database..."
    sqlite3 "$DB_FILE" < db/schema.sql
fi

echo "Starting Goodbuzz..."

exec ./goodbuzz