#!/usr/bin/env python3
import requests
import json

BASE_URL = "http://localhost:8080"

print("=== Testing Customer Support API ===\n")

# Test 1: Health Check
print("1. Health Check:")
try:
    resp = requests.get(f"{BASE_URL}/health")
    print(f"   Status: {resp.status_code}")
    print(f"   Response: {json.dumps(resp.json(), indent=2)}")
except Exception as e:
    print(f"   Error: {e}")
print()

# Test 2: Create Conversation
print("2. Create Conversation:")
try:
    payload = {
        "type": "natural_language",
        "initial_message": "I want to return my shoes"
    }
    resp = requests.post(f"{BASE_URL}/api/v1/conversations", json=payload)
    print(f"   Status: {resp.status_code}")
    print(f"   Response: {json.dumps(resp.json(), indent=2)}")
    
    if resp.status_code == 201:
        conv_id = resp.json().get("id")
        print(f"   ✓ Conversation created with ID: {conv_id}")
except Exception as e:
    print(f"   Error: {e}")
print()

# Test 3: Create Return Request
print("3. Create Return Request:")
try:
    payload = {
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
    }
    resp = requests.post(f"{BASE_URL}/api/v1/returns", json=payload)
    print(f"   Status: {resp.status_code}")
    print(f"   Response: {json.dumps(resp.json(), indent=2)}")
    
    if resp.status_code == 201:
        return_id = resp.json().get("id")
        return_ref = resp.json().get("return_reference")
        print(f"   ✓ Return request created with ID: {return_id}")
        print(f"   ✓ Return reference: {return_ref}")
except Exception as e:
    print(f"   Error: {e}")
print()

print("=== All tests completed ===")
