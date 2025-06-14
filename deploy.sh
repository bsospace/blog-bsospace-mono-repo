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

# Git stash to save current changes
git stash save "Saving current changes before deployment"
echo "Stashed current changes."

# Pull latest changes from the master branch
git pull origin master
echo "Pulled latest changes from main branch."

echo "Starting production containers with docker-compose.prod.yml..."

docker compose -f docker-compose.prod.yml up -d --build

echo "Deployment complete!"
