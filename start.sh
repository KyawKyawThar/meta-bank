#!/bin/sh
#in alpine image bashshell is not abailable

set -e

echo 'run db migration'
# shellcheck disable=SC2039
source /app/app.env

/app/db/migrate -path /app/db/migrations -database "$DB_SOURCE" -verbose up
echo "start the app"

# take all parameter passed to the script and run it
exec "$@"