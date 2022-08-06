#!/bin/bash

echo "START...";

if [[ $1 != "up" && $1 != "down" ]];then
  echo "ONLY UP/DOWN ALLOWED IN FIRST ARG"
  exit 1
fi

echo "MIGRATING AUTH";
migrate -database postgres://"$PG_AUTH_USER":"$PG_AUTH_USER"@pgs:5432/"$PG_DB"?sslmode=disable -path /migrations/alsiberij-api-auth/ "$1" "$2";
echo "DONE";
