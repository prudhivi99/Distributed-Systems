#!/bin/bash

# ============================================
# MinISys Infrastructure Clean Script
# WARNING: This removes all containers and data!
# ============================================

echo "⚠️  WARNING: This will remove all containers and data!"
read -p "Are you sure? (y/N): " confirm

if [ "$confirm" != "y" ] && [ "$confirm" != "Y" ]; then
    echo "Cancelled."
    exit 0
fi

echo ""
echo "��️  Removing MinISys Infrastructure..."

docker stop minisys-consul minisys-rabbitmq minisys-redis minisys-postgres 2>/dev/null
docker rm minisys-consul minisys-rabbitmq minisys-redis minisys-postgres 2>/dev/null

echo ""
echo "✅ All infrastructure removed"
