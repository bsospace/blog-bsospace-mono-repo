#!/bin/bash

set -e

echo "🔧 Combining frontend and backend .env files..."

# Clear previous .env if exists
rm -f .env

# Combine envs
cat frontend/.env > .env
echo "" >> .env
cat backend/.env >> .env

echo "Combined .env created successfully."

echo "Starting backend-related containers with docker-compose.dev.yml..."

docker compose -f docker-compose.dev.yml up -d --build

echo "Containers are up. Starting frontend dev server..."

cd frontend
pnpm install
pnpm dev
