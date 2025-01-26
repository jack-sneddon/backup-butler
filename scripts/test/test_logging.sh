#!/bin/bash

GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m'

CONFIG="config/examples/sync.yaml"

function test_log_level() {
   printf "Testing %-6s level... " "${1:-default}"
   
   cmd="./bin/backup-butler sync -c $CONFIG"
   [[ $1 != "default" ]] && cmd="$cmd --log-level $1"
   
   output=$($cmd 2>&1)
   
   case "$1" in
       "default"|"error")
           [[ ! $output =~ INFO && ! $output =~ DEBUG ]] && echo -e "${GREEN}PASS${NC}" || echo -e "${RED}FAIL${NC}"
           ;;
       "debug")
           [[ $output =~ DEBUG ]] && echo -e "${GREEN}PASS${NC}" || echo -e "${RED}FAIL${NC}"
           ;;
       "info")
           [[ $output =~ INFO && ! $output =~ DEBUG ]] && echo -e "${GREEN}PASS${NC}" || echo -e "${RED}FAIL${NC}"
           ;;
       "warn")
           [[ $output =~ WARN && ! $output =~ INFO && ! $output =~ DEBUG ]] && echo -e "${GREEN}PASS${NC}" || echo -e "${RED}FAIL${NC}"
           ;;
   esac
}

echo "Testing Log Levels..."
test_log_level "default"
test_log_level "debug"
test_log_level "info"
test_log_level "warn"
test_log_level "error"