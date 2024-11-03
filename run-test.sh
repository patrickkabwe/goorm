#!/bin/bash

set -e

if [ -t 1 ]; then
    RED='\033[0;31m'
    GREEN='\033[0;32m'
    YELLOW='\033[1;33m'
    NC='\033[0m'
else
    RED=''
    GREEN=''
    YELLOW=''
    NC=''
fi

log() {
    echo -e "[$(date '+%Y-%m-%d %H:%M:%S')] $1"
}

# Function to check if postgres is ready
wait_for_postgres() {
    if [ -z "$POSTGRES_DSN" ]; then
        log "${RED}Error: POSTGRES_DSN environment variable is not set${NC}"
        exit 1
    fi

    # Parse host and port from DSN
    local host=$(echo $POSTGRES_DSN | sed -n 's/.*@\([^:]*\):.*/\1/p')
    local port=$(echo $POSTGRES_DSN | sed -n 's/.*:\([0-9]*\)\/.*/\1/p')

    # Default to localhost:5432 if parsing fails
    host=${host:-localhost}
    port=${port:-9920}

    log "${YELLOW}Waiting for PostgreSQL to be ready at ${host}:${port}...${NC}"
    
    local retries=0
    local max_retries=30
    
    while ! nc -z $host $port; do
        retries=$((retries + 1))
        if [ $retries -eq $max_retries ]; then
            log "${RED}Failed to connect to PostgreSQL after $max_retries attempts${NC}"
            exit 1
        fi
        log "PostgreSQL is unavailable - sleeping (attempt $retries/$max_retries)"
        sleep 1
    done
    
    log "${GREEN}PostgreSQL is up and running!${NC}"
}

# Function to check if required commands exist
check_requirements() {
    local requirements=("go" "nc")
    local missing=()

    for cmd in "${requirements[@]}"; do
        if ! command -v "$cmd" &> /dev/null; then
            missing+=("$cmd")
        fi
    done

    if [ ${#missing[@]} -ne 0 ]; then
        log "${RED}Error: Required commands not found: ${missing[*]}${NC}"
        log "Please install the missing requirements and try again."
        exit 1
    fi
}

# Function to run database migrations
run_migrations() {
    log "${YELLOW}Running database migrations...${NC}"
    
    if [ -d "migrations" ]; then
        if command -v migrate &> /dev/null; then
            migrate -database "${POSTGRES_DSN}" -path migrations up
            return $?
        else
            log "${YELLOW}Migrate command not found, trying goorm-cli...${NC}"
        fi
    fi
    
    # Check if goorm-cli directory exists
    if [ -d "goorm-cli" ]; then
        log "${YELLOW}Running migrations using goorm-cli...${NC}"

        CURRENT_DIR=$(pwd)
        
        cd goorm-cli || {
            log "${RED}Failed to change to goorm-cli directory${NC}"
            return 1
        }
        
        if go run . push; then
            log "${GREEN}Successfully ran goorm-cli migrations${NC}"
            cd "$CURRENT_DIR"
            return 0
        else
            log "${RED}Failed to run goorm-cli migrations${NC}"
            cd "$CURRENT_DIR"
            return 1
        fi
    else
        log "${RED}Neither migrations directory nor goorm-cli directory found${NC}"
        return 1
    fi
}

# Function to run tests
run_tests() {
    log "${YELLOW}Running tests...${NC}"
    
    # Print DSN with masked password
    local masked_dsn=$(echo $POSTGRES_DSN | sed 's/:[^:]*@/:\*\*\*@/')
    log "Using DSN: $masked_dsn"
    
    # Run tests with coverage
    if go test -count=1 -v -cover ./...; then
        log "${GREEN}Tests completed successfully!${NC}"
    else
        log "${RED}Tests failed!${NC}"
        exit 1
    fi
}

# Main execution
main() {
    log "${GREEN}Starting test runner...${NC}"
    
    if [ -z "$POSTGRES_DSN" ]; then
        log "${RED}Error: POSTGRES_DSN environment variable is required${NC}"
        exit 1
    fi
    
    check_requirements
    wait_for_postgres
    run_migrations
    run_tests
    
    log "${GREEN}All operations completed successfully!${NC}"
}

# Run main function
main