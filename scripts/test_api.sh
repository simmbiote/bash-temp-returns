#!/bin/bash

echo "Testing API endpoints..."
echo ""

echo "1. Health Check:"
curl -s http://localhost:8080/health | jq
echo ""

echo "2. Creating a conversation:"
curl -s -X POST http://localhost:8080/api/v1/conversations \
  -H "Content-Type: application/json" \
  -d '{
    "type": "natural_language",
    "initial_message": "I want to return my shoes"
  }' | jq
echo ""

echo "3. Creating a return request:"
curl -s -X POST http://localhost:8080/api/v1/returns \
  -H "Content-Type: application/json" \
  -d '{
    "order_number": "ORD-123456",
    "items": [
      {
        "order_item_id": "ITEM-1",
        "product_id": "PROD-456",
        "quantity": 1
      }
    ],
    "reason": "wrong_size",
    "refund_method": "original_payment",
    "delivery_method": "collect",
    "collection_point": "Store 123"
  }' | jq
echo ""

echo "Done!"
