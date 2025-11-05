#!/bin/bash

# Customer Profiler Agent - Test Runner Script

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
BASE_URL="${BASE_URL:-http://localhost:8080}"
TEST_CLIENT="./bin/test-client"

# Print colored output
print_info() {
    echo -e "${BLUE}ℹ ${1}${NC}"
}

print_success() {
    echo -e "${GREEN}✓ ${1}${NC}"
}

print_error() {
    echo -e "${RED}✗ ${1}${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠ ${1}${NC}"
}

# Check if server is running
check_server() {
    print_info "Checking if server is running at ${BASE_URL}..."
    if curl -s -o /dev/null -w "%{http_code}" "${BASE_URL}/health" | grep -q "200"; then
        print_success "Server is running"
        return 0
    else
        print_error "Server is not running at ${BASE_URL}"
        return 1
    fi
}

# Build test client
build_test_client() {
    print_info "Building test client..."
    if [ ! -d "bin" ]; then
        mkdir -p bin
    fi
    
    if go build -o "${TEST_CLIENT}" cmd/test/main.go; then
        print_success "Test client built successfully"
        return 0
    else
        print_error "Failed to build test client"
        return 1
    fi
}

# Run tests
run_tests() {
    local test_type="${1:-all}"
    local idea="${2}"
    
    print_info "Running ${test_type} test(s)..."
    
    if [ "${test_type}" = "custom" ]; then
        if [ -z "${idea}" ]; then
            print_error "Business idea is required for custom test"
            echo "Usage: $0 custom \"Your business idea\""
            exit 1
        fi
        "${TEST_CLIENT}" -url "${BASE_URL}" -test custom -idea "${idea}"
    else
        "${TEST_CLIENT}" -url "${BASE_URL}" -test "${test_type}"
    fi
}

# Show help
show_help() {
    cat << EOF
Customer Profiler Agent - Test Runner

Usage: $0 [COMMAND] [OPTIONS]

Commands:
    all             Run all tests (default)
    health          Test health endpoint
    agent           Test agent card endpoint
    profile         Test profile generation with default idea
    custom "idea"   Test profile generation with custom business idea
    build           Build test client only
    check           Check if server is running
    help            Show this help message

Environment Variables:
    BASE_URL        Base URL of the agent (default: http://localhost:8080)

Examples:
    $0                                      # Run all tests
    $0 health                               # Test health endpoint only
    $0 custom "A mobile app for pet care"  # Test with custom idea
    BASE_URL=https://example.com $0 all    # Test remote server

Before running tests:
    1. Start the server: make run (or go run cmd/server/main.go)
    2. Set GEMINI_API_KEY environment variable
    3. Run tests: $0

EOF
}

# Main script
main() {
    local command="${1:-all}"
    
    case "${command}" in
        help|-h|--help)
            show_help
            exit 0
            ;;
        build)
            build_test_client
            exit $?
            ;;
        check)
            check_server
            exit $?
            ;;
        all|health|agent|profile)
            # Build test client if it doesn't exist
            if [ ! -f "${TEST_CLIENT}" ]; then
                build_test_client || exit 1
            fi
            
            # Check if server is running
            if ! check_server; then
                print_warning "Server is not running. Start it with: make run"
                exit 1
            fi
            
            # Map 'agent' to 'agent-card' for test client
            if [ "${command}" = "agent" ]; then
                run_tests "agent-card"
            else
                run_tests "${command}"
            fi
            ;;
        custom)
            local idea="${2}"
            
            # Build test client if it doesn't exist
            if [ ! -f "${TEST_CLIENT}" ]; then
                build_test_client || exit 1
            fi
            
            # Check if server is running
            if ! check_server; then
                print_warning "Server is not running. Start it with: make run"
                exit 1
            fi
            
            run_tests "custom" "${idea}"
            ;;
        *)
            print_error "Unknown command: ${command}"
            echo ""
            show_help
            exit 1
            ;;
    esac
}

# Run main function
main "$@"