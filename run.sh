#!/bin/bash

# Get the directory where the script is located
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

# Function to cleanup processes on exit
cleanup() {
    echo ""
    echo "Shutting down..."
    if [ ! -z "$BACKEND_PID" ]; then
        echo "Killing backend (PID: $BACKEND_PID)"
        kill $BACKEND_PID 2>/dev/null
    fi
    if [ ! -z "$FRONTEND_PID" ]; then
        echo "Killing frontend (PID: $FRONTEND_PID)"
        kill $FRONTEND_PID 2>/dev/null
    fi
    # Wait a bit for graceful shutdown
    sleep 1
    # Force kill if still running
    if [ ! -z "$BACKEND_PID" ] && kill -0 $BACKEND_PID 2>/dev/null; then
        kill -9 $BACKEND_PID 2>/dev/null
    fi
    if [ ! -z "$FRONTEND_PID" ] && kill -0 $FRONTEND_PID 2>/dev/null; then
        kill -9 $FRONTEND_PID 2>/dev/null
    fi
    echo "Done."
    exit 0
}

# Set up trap to catch Ctrl+C and cleanup
trap cleanup SIGINT SIGTERM

# Kill any existing instances on the ports
echo "Checking for existing instances..."

# Kill backend on port 59731
BACKEND_EXISTING=$(lsof -ti:59731 2>/dev/null)
if [ ! -z "$BACKEND_EXISTING" ]; then
    echo "Killing existing backend process(es) on port 59731: $BACKEND_EXISTING"
    kill $BACKEND_EXISTING 2>/dev/null
    sleep 1
    # Force kill if still running
    if lsof -ti:59731 >/dev/null 2>&1; then
        kill -9 $(lsof -ti:59731) 2>/dev/null
    fi
fi

# Kill frontend on port 53461
FRONTEND_EXISTING=$(lsof -ti:53461 2>/dev/null)
if [ ! -z "$FRONTEND_EXISTING" ]; then
    echo "Killing existing frontend process(es) on port 53461: $FRONTEND_EXISTING"
    kill $FRONTEND_EXISTING 2>/dev/null
    sleep 1
    # Force kill if still running
    if lsof -ti:53461 >/dev/null 2>&1; then
        kill -9 $(lsof -ti:53461) 2>/dev/null
    fi
fi

# Start backend
echo "Starting backend..."
go run ./cmd/waynebot &
BACKEND_PID=$!

# Wait a moment for backend to start
sleep 2

# Start frontend
echo "Starting frontend..."
cd frontend
npm run dev &
FRONTEND_PID=$!
cd ..

echo ""
echo "Backend running (PID: $BACKEND_PID) on port 59731"
echo "Frontend running (PID: $FRONTEND_PID) on port 53461"
echo "Visit http://localhost:53461"
echo ""
echo "Press Ctrl+C to stop both services"
echo ""

# Wait for both processes
wait $BACKEND_PID $FRONTEND_PID
