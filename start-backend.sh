#!/bin/bash
cd /mnt/c/universe/workspace/repo/subboard/backend
export APP_ENV=development
export APP_SECRET=dev-secret-change-in-production
export INIT_TOKEN=dev-init-token-12345
export DB_DRIVER=sqlite
export DB_DSN=submanager.db
export SUB_BASE_URL=http://localhost:8080
export ALLOW_REGISTER=true
export AGENT_REPORT_INTERVAL=60
export AGENT_OFFLINE_TIMEOUT=180

./submanager > backend.log 2>&1 &
BACKEND_PID=$!
echo "Backend PID: $BACKEND_PID"
echo $BACKEND_PID > backend.pid
sleep 3
if ps -p $BACKEND_PID > /dev/null 2>&1; then
    echo "Backend is running"
else
    echo "Backend failed to start"
    cat backend.log
fi