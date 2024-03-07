#!/bin/sh
set -e

echo "running db migration"
source /app/app.env
/app/migrate -path /app/migration -database "$DB_SOURCE" -verbose up

echo "starting the app"
exec "$@"