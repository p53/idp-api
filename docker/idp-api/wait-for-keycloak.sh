#!/bin/sh

set -e

CMD="$1"
shift
CMD_ARGS="$@"

LOOPS=10
until curl "${KEYCLOAK_PROTO}://${KEYCLOAK_HOST}:${KEYCLOAK_PORT}/auth/admin"; do
  >&2 echo "Keycloak is unavailable - sleeping"
  sleep 1
  if [ $LOOPS -eq 10 ]
  then
    break
  fi
done

sleep 30

>&2 echo "Keycloak is up - executing command"

exec $CMD $CMD_ARGS
