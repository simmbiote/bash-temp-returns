# API Quick Reference - v0.2.0

## New CRUD Endpoints

### Conversations

#### Get Conversation Details
```bash
GET /api/v1/conversations/:id
```

**Response**:
```json
{
  "id": "920d722f-b940-4bc4-8ef9-7544c4e652a0",
  "customer_id": "demo-customer-123",
  "type": "natural_language",
  "context": {},
  "messages": [],
  "created_at": "2026-02-04T14:24:41+02:00",
  "updated_at": "2026-02-04T14:24:41+02:00"
}
```

#### Get Conversation Messages
```bash
GET /api/v1/conversations/:id/messages
```

**Response**:
```json
{
  "messages": [
    {
      "id": "a3342d56-09c6-462c-b74e-e7d113db34fa",
      "role": "user",
      "content": "I want to return my shoes",
      "created_at": "2026-02-04T14:24:52+02:00"
    },
    {
      "id": "ebf03782-4e79-445f-9dae-6b1c42ad6939",
      "role": "assistant",
      "content": "Thank you for your message. How can I assist you further?",
      "created_at": "2026-02-04T14:24:52+02:00"
    }
  ]
}
```

#### Send Message to Conversation
```bash
POST /api/v1/conversations/:id/messages
Content-Type: application/json

{
  "content": "I want to return my shoes",
  "metadata": {
    "source": "web"
  }
}
```

**Response**:
```json
{
  "user_message": {
    "id": "a3342d56-09c6-462c-b74e-e7d113db34fa",
    "role": "user",
    "content": "I want to return my shoes",
    "created_at": "2026-02-04T14:24:52+02:00"
  },
  "assistant_message": {
    "id": "ebf03782-4e79-445f-9dae-6b1c42ad6939",
    "role": "assistant",
    "content": "Thank you for your message. How can I assist you further?",
    "created_at": "2026-02-04T14:24:52+02:00"
  }
}
```

### Returns

#### Get Return Request Details
```bash
GET /api/v1/returns/:id
```

**Response**:
```json
{
  "id": "d5df104a-e07e-4236-b87b-5622ecc16236",
  "conversation_id": "",
  "customer_id": "demo-customer-123",
  "order_number": "ORD-123456",
  "items": [
    {
      "order_item_id": "ITEM-1",
      "product_id": "PROD-456",
      "quantity": 1
    }
  ],
  "reason": "size_issue",
  "detailed_reason": "Shoes are too small",
  "refund_method": "original_payment",
  "delivery_method": "collection",
  "collection_point": "Cape Town Store",
  "status": "approved",
  "estimated_refund_date": "2026-02-11T14:24:15+02:00",
  "refund_amount_cents": 15999,
  "return_reference": "RET-d5df104a",
  "created_at": "2026-02-04T14:24:15+02:00",
  "updated_at": "2026-02-04T14:24:15+02:00"
}
```

#### List Customer Returns
```bash
GET /api/v1/returns?customer_id=demo-customer-123
```

**Response**:
```json
{
  "returns": [
    {
      "id": "d5df104a-e07e-4236-b87b-5622ecc16236",
      "customer_id": "demo-customer-123",
      "order_number": "ORD-123456",
      "items": [...],
      "status": "approved",
      "refund_amount_cents": 15999,
      "created_at": "2026-02-04T14:24:15+02:00"
    },
    {
      "id": "67cf7001-b7c9-417c-9815-6dbad86b1e44",
      "customer_id": "demo-customer-123",
      "order_number": "ORD-123456",
      "items": [...],
      "status": "approved",
      "refund_amount_cents": 17998,
      "created_at": "2026-02-04T14:25:03+02:00"
    }
  ]
}
```

## Orders API Integration

### Features
- ✅ **Order Validation**: Validates order exists and belongs to customer
- ✅ **Item Validation**: Validates items exist in order and are returnable
- ✅ **Refund Calculation**: Calculates actual refund amount based on item prices
- ✅ **Status Check**: Validates order is in returnable state (delivered)

### Create Return (Updated)
```bash
POST /api/v1/returns
Content-Type: application/json

{
  "customer_id": "demo-customer-123",
  "order_number": "ORD-123456",
  "items": [
    {
      "order_item_id": "ITEM-1",
      "product_id": "PROD-456",
      "quantity": 1
    }
  ],
  "reason": "size_issue",
  "detailed_reason": "Shoes are too small",
  "refund_method": "original_payment",
  "delivery_method": "collection",
  "collection_point": "Cape Town Store"
}
```

**Response** (now includes calculated refund):
```json
{
  "id": "d5df104a-e07e-4236-b87b-5622ecc16236",
  "return_reference": "RET-d5df104a",
  "status": "approved",
  "estimated_refund_date": "2026-02-11T14:24:15+02:00",
  "refund_amount_cents": 15999,
  "message": "Your return has been submitted successfully."
}
```

## Mock Data

### Available Test Orders

#### ORD-123456 (Delivered)
- **Items**:
  - `ITEM-1`: Nike Air Max 270 - $159.99
  - `ITEM-2`: Adidas Running Socks - $19.99
- **Total**: $179.98

#### ORD-789012 (Delivered)
- **Items**:
  - `ITEM-3`: Puma T-Shirt (x2) - $59.98

#### ORD-333444 (Shipped - NOT Returnable)
- **Items**:
  - `ITEM-4`: New Balance Sneakers - $129.99

### Test Scenarios

#### Scenario 1: Single Item Return
```bash
curl -X POST http://localhost:8080/api/v1/returns \
  -H "Content-Type: application/json" \
  -d '{
    "customer_id": "demo-customer-123",
    "order_number": "ORD-123456",
    "items": [{"order_item_id": "ITEM-1", "product_id": "PROD-456", "quantity": 1}],
    "reason": "size_issue",
    "refund_method": "original_payment",
    "delivery_method": "collection"
  }'
```
**Expected Refund**: $159.99

#### Scenario 2: Multiple Items Return
```bash
curl -X POST http://localhost:8080/api/v1/returns \
  -H "Content-Type: application/json" \
  -d '{
    "customer_id": "demo-customer-123",
    "order_number": "ORD-123456",
    "items": [
      {"order_item_id": "ITEM-1", "product_id": "PROD-456", "quantity": 1},
      {"order_item_id": "ITEM-2", "product_id": "PROD-789", "quantity": 3}
    ],
    "reason": "wrong_item",
    "refund_method": "original_payment",
    "delivery_method": "collection"
  }'
```
**Expected Refund**: $179.98

#### Scenario 3: Conversation Flow
```bash
# 1. Create conversation
curl -X POST http://localhost:8080/api/v1/conversations \
  -H "Content-Type: application/json" \
  -d '{"customer_id": "demo-customer-123", "type": "natural_language", "subject": "Return help"}'

# 2. Send message
curl -X POST http://localhost:8080/api/v1/conversations/{id}/messages \
  -H "Content-Type: application/json" \
  -d '{"content": "I want to return my shoes"}'

# 3. Get conversation details
curl http://localhost:8080/api/v1/conversations/{id}

# 4. Get all messages
curl http://localhost:8080/api/v1/conversations/{id}/messages
```

## Configuration

### Development (Mock)
```bash
# .env
ORDERS_API_URL=mock
PORT=8080
DB_PATH=./data/chatbot.db
LOG_LEVEL=debug
```

### Production
```bash
# .env
ORDERS_API_URL=https://api.orders.example.com
ORDERS_API_KEY=your-api-key-here
PORT=8080
DB_PATH=./data/chatbot.db
LOG_LEVEL=info
```

## Error Handling

### Order Not Found
```json
{
  "error": "order not found: ORD-999999"
}
```

### Invalid Order Items
```json
{
  "error": "item ITEM-999 not found in order ORD-123456"
}
```

### Order Not Returnable
```json
{
  "error": "item ITEM-4 cannot be returned (order status: shipped)"
}
```

### Conversation Completed
```json
{
  "error": "conversation is already completed"
}
```

## Quick Start

1. **Start Server**:
   ```bash
   ORDERS_API_URL=mock go run cmd/api/main.go
   ```

2. **Test Health**:
   ```bash
   curl http://localhost:8080/health
   ```

3. **Create Return with Refund Calculation**:
   ```bash
   curl -X POST http://localhost:8080/api/v1/returns \
     -H "Content-Type: application/json" \
     -d @test-return.json
   ```

4. **Get Return Details**:
   ```bash
   curl http://localhost:8080/api/v1/returns/{return_id}
   ```

5. **List Customer Returns**:
   ```bash
   curl "http://localhost:8080/api/v1/returns?customer_id=demo-customer-123"
   ```

## Changelog

### v0.2.0 (2026-02-04)
- ✅ Added Orders API integration
- ✅ Added refund calculation based on order items
- ✅ Added order and item validation
- ✅ Added GET /api/v1/returns/:id endpoint
- ✅ Added GET /api/v1/returns endpoint (list)
- ✅ Added GET /api/v1/conversations/:id endpoint
- ✅ Added GET /api/v1/conversations/:id/messages endpoint
- ✅ Added POST /api/v1/conversations/:id/messages endpoint
- ✅ Added mock Orders API client with sample data
- ✅ Added HTTP Orders API client for production

### v0.1.0 (Initial)
- POST /api/v1/conversations
- POST /api/v1/returns
- Basic conversation and return management
