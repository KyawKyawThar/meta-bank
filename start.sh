#!/bin/sh

set -e

echo 'run db migration'
# shellcheck disable=SC2039
source /app/app.env

/app/db/migrate -path /app/db/migrations -database "$DB_SOURCE_LOCAL" -verbose up
echo "start the app"

exec "$@"