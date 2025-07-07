#!/bin/sh
set -e

if [ -z "$DATABASE_URL" ]; then
    echo "DATABASE_URL not set" >&2
    exit 1
fi

for file in $(ls migrations/*.sql | sort); do
    echo "Applying $file"
    psql "$DATABASE_URL" -f "$file"
done
