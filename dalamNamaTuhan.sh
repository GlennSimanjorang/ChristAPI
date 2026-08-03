#!/bin/bash

# Setup & run ChristAPI with Docker (Bash)
# Usage:
#   ./dalamNamaTuhan.sh
#   ./dalamNamaTuhan.sh --no-build
#   ./dalamNamaTuhan.sh --no-build --no-migrate
#   ./dalamNamaTuhan.sh --migrate-only
#   ./dalamNamaTuhan.sh --restart
#   ./dalamNamaTuhan.sh --restart --service golang-christapi

set -e

# Parse arguments
NO_BUILD=false
NO_MIGRATE=false
RESTART=false
MIGRATE_ONLY=false
SERVICE=""

while [[ $# -gt 0 ]]; do
    case $1 in
        --no-build)
            NO_BUILD=true
            shift
            ;;
        --no-migrate)
            NO_MIGRATE=true
            shift
            ;;
        --restart)
            RESTART=true
            shift
            ;;
        --migrate-only)
            MIGRATE_ONLY=true
            shift
            ;;
        --service)
            SERVICE="$2"
            shift 2
            ;;
        *)
            echo "Unknown option: $1"
            exit 1
            ;;
    esac
done

echo -e "\033[36m[*] Bismillah... Starting ChristAPI setup...\033[0m"
echo ""

# 1. Check if Docker is running
echo -e "\033[33m[*] Checking Docker...\033[0m"
if ! docker ps > /dev/null 2>&1; then
    echo -e "\033[31m[ERROR] Docker is not running. Please start Docker and try again.\033[0m"
    exit 1
fi
echo -e "\033[32m[OK] Docker is running\033[0m"
echo ""

# 2. Check if .env.docker exists
echo -e "\033[33m[*] Checking Docker environment variables...\033[0m"
if [ ! -f ".env.docker" ]; then
    echo -e "\033[31m[ERROR] .env.docker not found\033[0m"
    exit 1
fi
echo -e "\033[32m[OK] .env.docker already exists\033[0m"
echo ""

# 3. Build Docker image
if [ "$MIGRATE_ONLY" = true ]; then
    echo -e "\033[33m[*] Skipping image build (--migrate-only)\033[0m"
elif [ "$NO_BUILD" = true ] || [ "$RESTART" = true ]; then
    echo -e "\033[33m[*] Skipping image build (--no-build or --restart)\033[0m"
else
    echo -e "\033[33m[*] Building Docker image...\033[0m"
    docker compose build --no-cache
    echo -e "\033[32m[OK] Build complete\033[0m"
fi
echo ""

# 4. Start or Restart services
if [ "$MIGRATE_ONLY" = true ]; then
    echo -e "\033[33m[*] Starting PostgreSQL only for migration...\033[0m"
    docker compose up -d postgres
    echo -e "\033[32m[OK] PostgreSQL started\033[0m"
    echo ""
elif [ "$RESTART" = true ]; then
    echo -e "\033[33m[*] Restart mode (--restart)\033[0m"
    if [ -z "$SERVICE" ]; then
        echo -e "\033[33m[*] Restarting all services...\033[0m"
        docker compose restart
    else
        echo -e "\033[33m[*] Restarting service: $SERVICE\033[0m"
        if ! docker compose restart "$SERVICE" 2>/dev/null; then
            echo -e "\033[33m[WARN] 'docker compose restart $SERVICE' failed, trying 'docker restart $SERVICE' (container-level)\033[0m"
            docker restart "$SERVICE"
        fi
    fi
    echo -e "\033[32m[OK] Services restarted\033[0m"
    echo ""
    echo -e "\033[33m[*] Skipping build/migrate in restart mode\033[0m"
else
    echo -e "\033[33m[*] Starting services (postgres, api)...\033[0m"
    docker compose down
    docker compose up -d
    echo -e "\033[32m[OK] Services started\033[0m"
    echo ""
fi

# 5. Wait for postgres to be healthy
echo -e "\033[33m[*] Waiting for PostgreSQL to be healthy...\033[0m"
sleep 3

for attempt in {1..15}; do
    if docker exec postgre-chrisapi pg_isready -U christ_user > /dev/null 2>&1; then
        echo -e "\033[32m[OK] PostgreSQL is healthy\033[0m"
        break
    fi
    echo "    Waiting... (attempt $attempt/15)"
    sleep 2
done
echo ""

# 6. Run migrations
if [ "$NO_MIGRATE" = true ] && [ "$MIGRATE_ONLY" = false ]; then
    echo -e "\033[33m[*] Skipping migrations (--no-migrate)\033[0m"
else
    echo -e "\033[33m[*] Running database migrations...\033[0m"
    docker compose run --rm migrate -path=/migrations -database "postgres://christ_user:christ_password@postgre-chrisapi:5432/christ_db?sslmode=disable" up
    echo -e "\033[32m[OK] Migrations complete\033[0m"
fi
echo ""

if [ "$MIGRATE_ONLY" = true ]; then
    echo -e "\033[33m[*] Migration-only mode selesai.\033[0m"
    echo -e "\033[32m[SUCCESS] Migration applied successfully.\033[0m"
    exit 0
fi

# 7. Show status
echo -e "\033[33m[*] Service status:\033[0m"
docker compose ps
echo ""

# 8. Show access info
echo -e "\033[32m========================================\033[0m"
echo -e "\033[32m[SUCCESS] ChristAPI is ready!\033[0m"
echo -e "\033[32m========================================\033[0m"
echo ""

echo -e "\033[36m[*] API Server:\033[0m"
echo -e "    http://localhost:3001"
echo ""

echo -e "\033[36m[*] Database:\033[0m"
echo -e "    Host: localhost"
echo -e "    Port: 5433"
echo -e "    Database: christ_db"
echo -e "    User: christ_user"
echo -e "    Password: christ_password"
echo ""

echo -e "\033[36m[*] Useful commands:\033[0m"
echo -e "    docker compose logs -f                 # View logs"
echo -e "    docker compose exec golang-christapi sh # Access API container"
echo -e "    docker compose down                    # Stop services"
echo ""

echo -e "\033[36m[*] DBeaver connection string:\033[0m"
echo -e "    postgres://christ_user:christ_password@localhost:5433/christ_db"
echo ""
