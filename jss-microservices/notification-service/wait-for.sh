#!/bin/sh

# This small script blocks until the given host and port become
# reachable.  It can be used in your Dockerfile CMD or entrypoint to
# wait for dependent services (like RabbitMQ or MongoDB) to be ready
# before starting the application.  When the host:port combination
# becomes available the script execs the given command.

set -e

HOST="$1"
PORT="$2"
shift 2
CMD="$@"

echo "Waiting for $HOST:$PORT..."

while ! nc -z "$HOST" "$PORT"; do
  sleep 2
done

echo "$HOST:$PORT is available, starting app..."
exec $CMD