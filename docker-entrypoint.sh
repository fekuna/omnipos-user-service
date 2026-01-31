#!/bin/sh
# Development entrypoint

echo "Starting user-service with Air..."
cd /app/omnipos-user-service
exec air -c .air.toml
