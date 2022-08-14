#!/bin/bash

echo "START...";
echo "MIGRATING AUTH";
migrate -database postgres://"$1":"$2"@"$PG_AUTH_HOST":"$PG_AUTH_PORT"/"$PG_AUTH_DB"?sslmode=disable -path /migrations/alsiberij-api-auth/ "$3" "$4" && echo "DONE";
