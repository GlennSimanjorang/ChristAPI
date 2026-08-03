#!/bin/bash

# Reset ChristAPI Database
# Usage:
#   ./resetDatabase.sh                    # Interactive menu
#   ./resetDatabase.sh truncate           # Clear all tables (keep schema)
#   ./resetDatabase.sh drop               # Drop & recreate database
#   ./resetDatabase.sh drop --migrate     # Drop, recreate & run migrations

set -e

CONTAINER="postgre-chrisapi"
DB_USER="christ_user"
DB_NAME="christ_db"
PASSWORD="christ_password"

echo -e "\033[36m[*] ChristAPI Database Reset Tool\033[0m"
echo ""

# Check if Docker is running
if ! docker ps > /dev/null 2>&1; then
    echo -e "\033[31m[ERROR] Docker is not running\033[0m"
    exit 1
fi

# Check if PostgreSQL container is running
if ! docker ps | grep -q "$CONTAINER"; then
    echo -e "\033[31m[ERROR] PostgreSQL container '$CONTAINER' is not running\033[0m"
    echo -e "\033[33m[*] Start it with: docker compose up -d postgres\033[0m"
    exit 1
fi

MODE="${1:-interactive}"
DO_MIGRATE="${2}"

# Interactive menu
if [ "$MODE" = "interactive" ]; then
    echo "Pilih opsi:"
    echo "  1) Truncate all tables (clear data, keep schema)"
    echo "  2) Drop & recreate database"
    echo "  3) Drop, recreate & run migrations"
    echo "  4) Exit"
    echo ""
    read -p "Pilihan [1-4]: " choice
    
    case $choice in
        1) MODE="truncate" ;;
        2) MODE="drop" ;;
        3) MODE="drop"; DO_MIGRATE="--migrate" ;;
        4) echo "Dibatalkan."; exit 0 ;;
        *) echo "Pilihan tidak valid"; exit 1 ;;
    esac
fi

# Truncate mode: clear all tables
if [ "$MODE" = "truncate" ]; then
    echo -e "\033[33m[*] Truncating all tables...\033[0m"
    docker exec "$CONTAINER" psql -U "$DB_USER" -d "$DB_NAME" -c \
        "TRUNCATE TABLE contacts, roles, user_otps, users, sites, news, user_points_ledger CASCADE;"
    
    if [ $? -eq 0 ]; then
        echo -e "\033[32m[OK] All tables truncated successfully\033[0m"
        echo -e "\033[36m[*] Database schema is intact, data has been cleared\033[0m"
    else
        echo -e "\033[31m[ERROR] Failed to truncate tables\033[0m"
        exit 1
    fi

# Drop mode: drop and recreate database
elif [ "$MODE" = "drop" ]; then
    echo -e "\033[33m[*] Dropping database '$DB_NAME'...\033[0m"
    docker exec "$CONTAINER" psql -U "$DB_USER" -c "DROP DATABASE IF EXISTS $DB_NAME;" 2>/dev/null
    
    echo -e "\033[33m[*] Creating new database '$DB_NAME'...\033[0m"
    docker exec "$CONTAINER" psql -U "$DB_USER" -c "CREATE DATABASE $DB_NAME;" 2>/dev/null
    
    if [ $? -eq 0 ]; then
        echo -e "\033[32m[OK] Database dropped and recreated successfully\033[0m"
    else
        echo -e "\033[31m[ERROR] Failed to drop/recreate database\033[0m"
        exit 1
    fi
    
    # Run migrations if requested
    if [ "$DO_MIGRATE" = "--migrate" ]; then
        echo ""
        echo -e "\033[33m[*] Running migrations...\033[0m"
        docker compose run --rm migrate -path=/migrations -database \
            "postgres://$DB_USER:$PASSWORD@$CONTAINER:5432/$DB_NAME?sslmode=disable" up
        
        if [ $? -eq 0 ]; then
            echo -e "\033[32m[OK] Migrations completed successfully\033[0m"
        else
            echo -e "\033[31m[ERROR] Migrations failed\033[0m"
            exit 1
        fi
    fi

else
    echo -e "\033[31m[ERROR] Invalid mode: $MODE\033[0m"
    echo "Usage: ./resetDatabase.sh [truncate|drop|interactive]"
    exit 1
fi

echo ""
echo -e "\033[32m========================================\033[0m"
echo -e "\033[32m[SUCCESS] Database reset completed!\033[0m"
echo -e "\033[32m========================================\033[0m"
