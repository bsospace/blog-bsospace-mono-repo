#!/bin/bash

set -e

echo "Combining frontend and backend .env files..."

# Check and combine .env files
if [[ -f frontend/.env && -f backend/.env ]]; then
  rm -f .env
  cat frontend/.env > .env
  echo "" >> .env
  cat backend/.env >> .env
  echo "Combined .env created successfully."
else
  echo "One or both .env files not found in frontend/ or backend/. Aborting."
  exit 1
fi

echo "Starting backend containers from docker-compose.dev.yml..."
docker compose -f docker-compose.dev.yml up -d --build

echo "Starting frontend development server..."
cd frontend
pnpm install
pnpm dev
