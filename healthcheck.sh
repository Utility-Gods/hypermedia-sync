#!/bin/sh
# Health check script that uses the PORT environment variable
PORT=${PORT:-8080}
wget --no-verbose --tries=1 -O- "http://127.0.0.1:${PORT}/health" > /dev/null 2>&1 || exit 1
