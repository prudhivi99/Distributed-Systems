#!/bin/bash

# ============================================
# MinISys Infrastructure Startup Script
# ============================================
# This script starts all required infrastructure
# services for the minisys-go project
# ============================================

set -e

echo "🚀 Starting MinISys Infrastructure..."
echo ""

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Function to check if container exists
container_exists() {
    docker ps -a --format '{{.Names}}' | grep -q "^$1$"
}

# Function to check if container is running
container_running() {
    docker ps --format '{{.Names}}' | grep -q "^$1$"
}

# Function to start or create container
start_container() {
    local name=$1
    local image=$2
    shift 2
    local args="$@"

    if container_running "$name"; then
        echo -e "${GREEN}✅ $name is already running${NC}"
    elif container_exists "$name"; then
        echo -e "${YELLOW}🔄 Starting existing $name container...${NC}"
        docker start "$name"
        echo -e "${GREEN}✅ $name started${NC}"
    else
        echo -e "${YELLOW}📦 Creating and starting $name...${NC}"
        eval "docker run -d --name $name $args $image"
        echo -e "${GREEN}✅ $name created and started${NC}"
    fi
}

echo "============================================"
echo "1. PostgreSQL (Database)"
echo "============================================"
start_container "minisys-postgres" "postgres:16-alpine" \
    "-e POSTGRES_USER=minisys \
     -e POSTGRES_PASSWORD=minisys123 \
     -e POSTGRES_DB=minisys \
     -p 5432:5432"

echo ""
echo "============================================"
echo "2. Redis (Cache)"
echo "============================================"
start_container "minisys-redis" "redis:7-alpine" \
    "-p 6379:6379"

echo ""
echo "============================================"
echo "3. RabbitMQ (Message Queue)"
echo "============================================"
start_container "minisys-rabbitmq" "rabbitmq:3-management-alpine" \
    "-p 5672:5672 \
     -p 15672:15672 \
     -e RABBITMQ_DEFAULT_USER=guest \
     -e RABBITMQ_DEFAULT_PASS=guest"

echo ""
echo "============================================"
echo "4. Consul (Service Discovery)"
echo "============================================"
start_container "minisys-consul" "consul:1.15" \
    "-p 8500:8500 \
     -p 8600:8600/udp \
     agent -server -ui -node=server-1 -bootstrap-expect=1 -client=0.0.0.0"

echo ""
echo "============================================"
echo "⏳ Waiting for services to be ready..."
echo "============================================"
sleep 10

echo ""
echo "============================================"
echo "5. Database Migration"
echo "============================================"

# Check if tables exist
TABLES_EXIST=$(docker exec minisys-postgres psql -U minisys -d minisys -t -c "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public';" 2>/dev/null | tr -d ' ')

if [ "$TABLES_EXIST" -gt "0" ] 2>/dev/null; then
    echo -e "${GREEN}✅ Database tables already exist${NC}"
else
    echo -e "${YELLOW}📦 Running database migrations...${NC}"
    
    # Create tables
    docker exec -i minisys-postgres psql -U minisys -d minisys << 'EOSQL'
-- Products table
CREATE TABLE IF NOT EXISTS products (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    price DECIMAL(10,2) NOT NULL,
    quantity INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Orders table
CREATE TABLE IF NOT EXISTS orders (
    id SERIAL PRIMARY KEY,
    customer_name VARCHAR(255) NOT NULL,
    total_amount DECIMAL(10,2) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Order items table
CREATE TABLE IF NOT EXISTS order_items (
    id SERIAL PRIMARY KEY,
    order_id INT REFERENCES orders(id),
    product_id INT NOT NULL,
    product_name VARCHAR(255) NOT NULL,
    quantity INT NOT NULL,
    price DECIMAL(10,2) NOT NULL
);

-- Insert sample products if empty
INSERT INTO products (name, price, quantity)
SELECT 'Laptop', 999.99, 10
WHERE NOT EXISTS (SELECT 1 FROM products WHERE name = 'Laptop');

INSERT INTO products (name, price, quantity)
SELECT 'Mouse', 29.99, 50
WHERE NOT EXISTS (SELECT 1 FROM products WHERE name = 'Mouse');

INSERT INTO products (name, price, quantity)
SELECT 'Keyboard', 79.99, 30
WHERE NOT EXISTS (SELECT 1 FROM products WHERE name = 'Keyboard');
EOSQL

    echo -e "${GREEN}✅ Database migrations completed${NC}"
fi

echo ""
echo "============================================"
echo "📊 Infrastructure Status"
echo "============================================"
docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}" | grep minisys

echo ""
echo "============================================"
echo "🎉 Infrastructure Ready!"
echo "============================================"
echo ""
echo "Services:"
echo "  PostgreSQL:  localhost:5432 (minisys/minisys123)"
echo "  Redis:       localhost:6379"
echo "  RabbitMQ:    localhost:5672 (guest/guest)"
echo "  RabbitMQ UI: http://localhost:15672"
echo "  Consul UI:   http://localhost:8500"
echo ""
echo "Next steps:"
echo "  1. Terminal 1: go run cmd/product-service/main.go"
echo "  2. Terminal 2: go run cmd/order-service/main.go"
echo "  3. Terminal 3: go run cmd/api-gateway/main.go"
echo ""
