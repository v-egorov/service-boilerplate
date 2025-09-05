#!/bin/bash

# Development helper script
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

cd "$PROJECT_ROOT"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

# Check if Docker is running
check_docker() {
    if ! docker info >/dev/null 2>&1; then
        print_error "Docker is not running. Please start Docker first."
        exit 1
    fi
}

# Main development menu
show_menu() {
    echo "🚀 Golang Service Boilerplate - Development Tools"
    echo "=================================================="
    echo ""
    echo "1. Start all services (Docker)"
    echo "2. Start development with hot reload"
    echo "3. View service logs"
    echo "4. Stop all services"
    echo "5. Run database migrations"
    echo "6. Test API endpoints"
    echo "7. Clean up (remove containers and volumes)"
    echo "8. Show service status"
    echo "9. Exit"
    echo ""
}

# Test API endpoints
test_api() {
    print_info "Testing API endpoints..."

    # Wait for services to be ready
    print_info "Waiting for services to be ready..."
    sleep 5

    # Test health endpoints
    if curl -s http://localhost:8080/health >/dev/null; then
        print_success "API Gateway health check passed"
    else
        print_error "API Gateway health check failed"
    fi

    if curl -s http://localhost:8081/health >/dev/null; then
        print_success "User Service health check passed"
    else
        print_error "User Service health check failed"
    fi

    # Test user creation
    print_info "Testing user creation..."
    response=$(curl -s -X POST http://localhost:8080/api/v1/users \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer test-token" \
        -d '{"email": "test@example.com", "first_name": "Test", "last_name": "User"}')

    if echo "$response" | grep -q "id"; then
        print_success "User creation test passed"
    else
        print_error "User creation test failed: $response"
    fi
}

# Show service status
show_status() {
    print_info "Service Status:"
    echo ""

    # Check Docker containers
    if docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}" | grep -q "service-boilerplate"; then
        echo "Docker Containers:"
        docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}" | grep "service-boilerplate"
        echo ""
    else
        print_warning "No service containers running"
    fi

    # Check ports
    echo "Port Status:"
    for port in 8080 8081 5432; do
        if ss -tulpn | grep -q ":$port "; then
            echo "  Port $port: ✅ In use"
        else
            echo "  Port $port: ❌ Free"
        fi
    done
}

# Main script logic
case "${1:-}" in
    "test-api")
        check_docker
        test_api
        ;;
    "status")
        show_status
        ;;
    "clean")
        print_info "Cleaning up containers and volumes..."
        docker-compose -f docker/docker-compose.yml down -v --remove-orphans
        print_success "Cleanup completed"
        ;;
    *)
        while true; do
            show_menu
            read -p "Choose an option (1-9): " choice

            case $choice in
                1)
                    check_docker
                    print_info "Starting all services..."
                    make up
                    ;;
                2)
                    check_docker
                    print_info "Starting development with hot reload..."
                    make dev
                    ;;
                3)
                    check_docker
                    print_info "Showing service logs..."
                    make logs
                    ;;
                4)
                    check_docker
                    print_info "Stopping all services..."
                    make down
                    ;;
                5)
                    print_info "Running database migrations..."
                    export DATABASE_URL="postgres://postgres:postgres@localhost:5432/service_db?sslmode=disable"
                    ~/go/bin/migrate -path services/user-service/migrations -database "$DATABASE_URL" up
                    print_success "Migrations completed"
                    ;;
                6)
                    check_docker
                    test_api
                    ;;
                7)
                    check_docker
                    print_info "Cleaning up containers and volumes..."
                    docker-compose -f docker/docker-compose.yml down -v --remove-orphans
                    print_success "Cleanup completed"
                    ;;
                8)
                    show_status
                    ;;
                9)
                    print_info "Goodbye! 👋"
                    exit 0
                    ;;
                *)
                    print_error "Invalid option. Please choose 1-9."
                    ;;
            esac

            echo ""
            read -p "Press Enter to continue..."
            clear
        done
        ;;
esac