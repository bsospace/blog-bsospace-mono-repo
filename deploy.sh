#!/bin/bash

set -e

echo "Combining frontend and backend .env files..."

# Clear previous .env if exists
rm -f .env

# Combine envs
cat frontend/.env > .env
echo "" >> .env
cat backend/.env >> .env

echo "Combined .env created successfully."

echo "Starting production containers with docker-compose.prod.yml..."

docker compose -f docker-compose.prod.yml up -d --build

echo "Deployment complete!"
