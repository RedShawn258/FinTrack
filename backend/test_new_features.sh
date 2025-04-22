#!/bin/bash

# Run tests for the new features
cd "$(dirname "$0")"

echo "Running tests for new backend features..."

# Test forecast feature
echo "Testing Forecast Feature..."
go test -v ./internal/handlers_test -run TestForecastExpensesHandlerBasic
go test -v ./internal/handlers_test -run TestForecastExpensesHandler_InvalidRequest

# Test gamification helper functions
echo "Testing Gamification Helper Functions..."
go test -v ./internal/handlers_test -run TestCalculateLevel
go test -v ./internal/handlers_test -run TestCalculatePointsToNextLevel
go test -v ./internal/handlers_test -run TestGetLevelTitle
go test -v ./internal/handlers_test -run TestGamificationHandlerWithNilDB

# Test stub endpoints
echo "Testing Stub Endpoints..."
go test -v ./internal/handlers_test -run TestAnalyticsHandler
go test -v ./internal/handlers_test -run TestNotificationHandler

echo "All tests completed!" 