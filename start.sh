#!/bin/sh

set -e

echo 'run db migration'
/app/db/migrate -path /app/db/migrations -database "$DB_SOURCE" -verbose up
echo "start the app"

exec "$@"