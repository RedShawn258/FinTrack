#!/bin/bash

# FinTrack Demo Script
# This script demonstrates the key features of the financial ledger API

set -e

API_URL="${API_URL:-http://localhost:8080}"
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  FinTrack Financial Ledger API Demo${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# Check if server is running
echo -e "${YELLOW}Checking if server is running...${NC}"
if ! curl -s "${API_URL}/health" > /dev/null; then
    echo -e "${YELLOW}⚠️  Server not running at ${API_URL}${NC}"
    echo -e "${YELLOW}Please start the server first:${NC}"
    echo "  cd backend && go run cmd/server/main.go"
    exit 1
fi

echo -e "${GREEN}✅ Server is running!${NC}"
echo ""

# Step 1: Create accounts
echo -e "${BLUE}Step 1: Creating accounts...${NC}"
echo ""

FROM_ACCOUNT_RESPONSE=$(curl -s -X POST "${API_URL}/accounts" \
  -H "Content-Type: application/json" \
  -d '{"initialBalance": 1000.0}')

FROM_ACCOUNT_ID=$(echo $FROM_ACCOUNT_RESPONSE | grep -o '"id":[0-9]*' | head -1 | cut -d':' -f2)

echo "Created Account 1:"
echo "$FROM_ACCOUNT_RESPONSE" | python3 -m json.tool 2>/dev/null || echo "$FROM_ACCOUNT_RESPONSE"
echo ""

TO_ACCOUNT_RESPONSE=$(curl -s -X POST "${API_URL}/accounts" \
  -H "Content-Type: application/json" \
  -d '{"initialBalance": 500.0}')

TO_ACCOUNT_ID=$(echo $TO_ACCOUNT_RESPONSE | grep -o '"id":[0-9]*' | head -1 | cut -d':' -f2)

echo "Created Account 2:"
echo "$TO_ACCOUNT_RESPONSE" | python3 -m json.tool 2>/dev/null || echo "$TO_ACCOUNT_RESPONSE"
echo ""

if [ -z "$FROM_ACCOUNT_ID" ] || [ -z "$TO_ACCOUNT_ID" ]; then
    echo -e "${YELLOW}⚠️  Could not extract account IDs. Using defaults.${NC}"
    FROM_ACCOUNT_ID=1
    TO_ACCOUNT_ID=2
fi

# Step 2: Check initial balances
echo -e "${BLUE}Step 2: Checking initial balances...${NC}"
echo ""

echo "Account ${FROM_ACCOUNT_ID} balance:"
curl -s "${API_URL}/accounts/${FROM_ACCOUNT_ID}/balance" | python3 -m json.tool 2>/dev/null || curl -s "${API_URL}/accounts/${FROM_ACCOUNT_ID}/balance"
echo ""

echo "Account ${TO_ACCOUNT_ID} balance:"
curl -s "${API_URL}/accounts/${TO_ACCOUNT_ID}/balance" | python3 -m json.tool 2>/dev/null || curl -s "${API_URL}/accounts/${TO_ACCOUNT_ID}/balance"
echo ""

# Step 3: Create transfer (first time)
echo -e "${BLUE}Step 3: Creating transfer (Idempotency-Key: demo-001)...${NC}"
echo ""

TRANSFER_RESPONSE=$(curl -s -X POST "${API_URL}/transfers" \
  -H "Content-Type: application/json" \
  -H "Idempotency-Key: demo-001" \
  -d "{
    \"fromAccountId\": ${FROM_ACCOUNT_ID},
    \"toAccountId\": ${TO_ACCOUNT_ID},
    \"amount\": 100.0,
    \"description\": \"Demo transfer\"
  }")

echo "Transfer Response:"
echo "$TRANSFER_RESPONSE" | python3 -m json.tool 2>/dev/null || echo "$TRANSFER_RESPONSE"
echo ""

# Step 4: Retry with same idempotency key (demonstrates idempotency)
echo -e "${BLUE}Step 4: Retrying with same Idempotency-Key (should return cached response)...${NC}"
echo ""

RETRY_RESPONSE=$(curl -s -X POST "${API_URL}/transfers" \
  -H "Content-Type: application/json" \
  -H "Idempotency-Key: demo-001" \
  -d "{
    \"fromAccountId\": ${FROM_ACCOUNT_ID},
    \"toAccountId\": ${TO_ACCOUNT_ID},
    \"amount\": 100.0,
    \"description\": \"Demo transfer\"
  }")

echo "Retry Response (should be cached):"
echo "$RETRY_RESPONSE" | python3 -m json.tool 2>/dev/null || echo "$RETRY_RESPONSE"
echo ""

# Step 5: Check updated balances
echo -e "${BLUE}Step 5: Checking updated balances...${NC}"
echo ""

echo "Account ${FROM_ACCOUNT_ID} balance (should be 900.0):"
curl -s "${API_URL}/accounts/${FROM_ACCOUNT_ID}/balance" | python3 -m json.tool 2>/dev/null || curl -s "${API_URL}/accounts/${FROM_ACCOUNT_ID}/balance"
echo ""

echo "Account ${TO_ACCOUNT_ID} balance (should be 600.0):"
curl -s "${API_URL}/accounts/${TO_ACCOUNT_ID}/balance" | python3 -m json.tool 2>/dev/null || curl -s "${API_URL}/accounts/${TO_ACCOUNT_ID}/balance"
echo ""

# Step 6: View ledger
echo -e "${BLUE}Step 6: Viewing account ledger...${NC}"
echo ""

echo "Account ${FROM_ACCOUNT_ID} ledger:"
curl -s "${API_URL}/accounts/${FROM_ACCOUNT_ID}/ledger?limit=10" | python3 -m json.tool 2>/dev/null || curl -s "${API_URL}/accounts/${FROM_ACCOUNT_ID}/ledger?limit=10"
echo ""

# Step 7: Check metrics
echo -e "${BLUE}Step 7: Checking Prometheus metrics...${NC}"
echo ""

echo "Outbox metrics:"
curl -s "${API_URL}/metrics" | grep -E "(outbox|reconciliation)" | head -10 || echo "No metrics found (may need to wait for dispatcher to process)"
echo ""

# Step 8: Health check
echo -e "${BLUE}Step 8: Health check...${NC}"
echo ""

echo "Health:"
curl -s "${API_URL}/health" | python3 -m json.tool 2>/dev/null || curl -s "${API_URL}/health"
echo ""

echo "Readiness:"
curl -s "${API_URL}/health/ready" | python3 -m json.tool 2>/dev/null || curl -s "${API_URL}/health/ready"
echo ""

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}  Demo Complete!${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo "Key Features Demonstrated:"
echo "  ✅ Account creation"
echo "  ✅ Transfer with idempotency"
echo "  ✅ Idempotent retry (cached response)"
echo "  ✅ Balance updates"
echo "  ✅ Ledger view"
echo "  ✅ Health checks"
echo ""
echo "View Swagger docs at: ${API_URL}/docs"
echo "View metrics at: ${API_URL}/metrics"
