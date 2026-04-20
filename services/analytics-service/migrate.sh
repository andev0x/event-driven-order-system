#!/bin/sh
set -e

MYSQL_OPTS="--skip-ssl --skip-ssl-verify-server-cert"
MYSQL_BIN="$(command -v mariadb || command -v mysql)"

log_json() {
    level="$1"
    message="$2"
    printf '{"level":"%s","service":"analytics-service","component":"migrations","message":"%s"}\n' "$level" "$message"
}

if [ -z "$MYSQL_BIN" ]; then
    log_json "error" "MySQL client binary not found"
    exit 1
fi

log_json "info" "Waiting for MySQL ($DB_HOST:$DB_PORT) to be ready"
until "$MYSQL_BIN" $MYSQL_OPTS -h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" -p"$DB_PASSWORD" "$DB_NAME" -e "SELECT 1" > /dev/null 2>&1; do
    sleep 2
done

log_json "info" "Running migrations"
for file in /root/migrations/*.sql; do
    if [ -f "$file" ]; then
        log_json "info" "Executing $file"
        "$MYSQL_BIN" $MYSQL_OPTS -h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" -p"$DB_PASSWORD" "$DB_NAME" < "$file"
    fi
done
log_json "info" "Migrations completed successfully"
