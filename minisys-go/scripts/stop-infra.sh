#!/bin/bash

# ============================================
# MinISys Infrastructure Stop Script
# ============================================

echo "🛑 Stopping MinISys Infrastructure..."
echo ""

docker stop minisys-consul minisys-rabbitmq minisys-redis minisys-postgres 2>/dev/null

echo ""
echo "✅ All infrastructure stopped"
echo ""
echo "To restart: ./scripts/start-infra.sh"
