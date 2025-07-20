#!/bin/bash
set -e

echo "Starting deployment..."

# Ensure same project name across environments (Jenkins + local)
export COMPOSE_PROJECT_NAME=blog-bsospace

# Clean up previous combined .env file
if [ -f .env ]; then
  echo "Removing previous combined .env..."
  rm .env
fi

# Combine frontend and backend .env files
echo "Combining frontend/.env and backend/.env into root .env..."
if [ -f frontend/.env ] && [ -f backend/.env ]; then
  cat frontend/.env > .env
  echo "" >> .env
  cat backend/.env >> .env
  echo "Combined .env created successfully."
else
  echo "Missing one or both .env files. Make sure frontend/.env and backend/.env exist."
  exit 1
fi

# Stash current changes before pulling
echo "Stashing local changes..."
git stash push -m "Auto-stash before deployment"

# Pull latest changes from remote
echo "Pulling latest changes from master branch..."
git pull origin master

# Start production containers
echo "Starting production containers with docker-compose.prod.yml..."
docker compose -f docker-compose.prod.yml up -d --build

echo "✅ Deployment complete!"
