#!/bin/bash

# Run all tests for the backend including existing and new ones
cd "$(dirname "$0")"

echo "======================================================================================"
echo "Running all FinTrack backend tests (including existing and new features)"
echo "======================================================================================"

# Set colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Count tests and failures
total_tests=0
passed_tests=0
failed_packages=0

# Function to run tests in a package and track results
run_tests() {
  local package=$1
  local test_pattern=$2
  
  echo -e "\n${YELLOW}Running tests for $package $test_pattern${NC}"
  
  if [ -z "$test_pattern" ]; then
    go test -v "./$package"
  else
    go test -v "./$package" -run "$test_pattern"
  fi
  
  local result=$?
  
  if [ $result -eq 0 ]; then
    echo -e "${GREEN}✓ Tests passed${NC}"
    ((passed_tests++))
  else
    echo -e "${RED}✗ Tests failed${NC}"
    ((failed_packages++))
  fi
  
  ((total_tests++))
  
  return $result
}

# Define packages to test
packages=(
  "internal/handlers_test"
  "internal/db"
  "internal/routes"
  "internal/middlewares"
  "internal/utils"
)

# Run general tests for each package
for pkg in "${packages[@]}"; do
  if [ -d "$pkg" ]; then
    run_tests "$pkg" ""
  else
    echo -e "${YELLOW}Package $pkg not found, skipping...${NC}"
  fi
done

# Run specific tests for new features
echo -e "\n${YELLOW}Running specific tests for new features${NC}"

# Forecast feature tests
run_tests "internal/handlers_test" "TestForecastExpensesHandler"

# Gamification feature tests
run_tests "internal/handlers_test" "TestGamification|TestCalculateLevel|TestCalculatePointsToNextLevel|TestGetLevelTitle"

# Stub endpoint tests
run_tests "internal/handlers_test" "TestAnalyticsHandler|TestNotificationHandler"

# Print summary
echo -e "\n======================================================================================"
echo -e "Test summary:"
echo -e "  Total test packages: $total_tests"
echo -e "  Passed: $passed_tests"
echo -e "  Failed: $failed_packages"
echo -e "======================================================================================"

# Return non-zero exit code if any tests failed
if [ $failed_packages -gt 0 ]; then
  echo -e "${RED}Some tests failed${NC}"
  exit 1
else
  echo -e "${GREEN}All tests passed${NC}"
  exit 0
fi 