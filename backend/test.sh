#!/bin/bash

# Run tests with verbose output
cd $(dirname $0)
go test -v ./internal/handlers_test/...

# Exit with the same status code as the test command
exit $? 