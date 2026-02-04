# API Endpoints Reference

## Overview

The Customer Support Chatbot API exposes RESTful HTTP endpoints for conversation management and returns processing. All endpoints follow REST principles with JSON request/response bodies.

## Base Configuration

**Base URL**: `http://localhost:8080` (development)  
**API Version**: v1  
**API Prefix**: `/api/v1`  
**Port**: 8080 (configurable via `PORT` environment variable)

## Authentication

**Current Status**: Demo mode in MVP  
**Customer ID**: Defaults to `demo-customer-123` when not authenticated  
**Future**: JWT Bearer token authentication

**Planned Authentication Flow**:
```
Authorization: Bearer <jwt-token>
```

Token will contain customer ID, extracted by authentication middleware.

## Content Type

**Request**: `Content-Type: application/json`  
**Response**: `Content-Type: application/json`

## Error Response Format

All errors return consistent JSON structure:

```json
{
  "error": "Descriptive error message"
}
```

**HTTP Status Codes**:
- `200 OK`: Successful GET request
- `201 Created`: Successful POST request (resource created)
- `400 Bad Request`: Invalid input or validation failure
- `404 Not Found`: Resource not found
- `500 Internal Server Error`: Server-side error

## Health and Status Endpoints

### Health Check

**Endpoint**: `GET /health`  
**Purpose**: Service health and status verification  
**Authentication**: None required

**Response**: `200 OK`
```json
{
  "status": "ok",
  "database": "sqlite",
  "db_status": "connected",
  "cache": "memory",
  "version": "0.1.0"
}
```

**Response Fields**:
- `status`: Overall service status (`ok`)
- `database`: Database type (`sqlite`)
- `db_status`: Database connection status (`connected` or `disconnected`)
- `cache`: Cache type (`memory`)
- `version`: API version

**Use Cases**:
- Health monitoring
- Load balancer health checks
- Service availability verification

---

### Ping

**Endpoint**: `GET /api/v1/ping`  
**Purpose**: API availability test  
**Authentication**: None required

**Response**: `200 OK`
```json
{
  "message": "pong",
  "timestamp": "2026-02-04T12:00:00Z"
}
```

**Use Cases**:
- Verify API is responding
- Network connectivity test
- Latency measurement

---

## Conversation Endpoints

### Create Conversation

**Endpoint**: `POST /api/v1/conversations`  
**Purpose**: Start a new customer support conversation  
**Authentication**: Customer ID (demo mode in MVP)

**Request Body**:
```json
{
  "type": "natural_language",
  "initial_message": "I need help with my order"
}
```

**Request Fields**:
- `type` (required): Conversation type (`natural_language` or `guided_flow`)
- `initial_message` (optional): Customer's opening message

**Response**: `201 Created`
```json
{
  "conversation_id": "550e8400-e29b-41d4-a716-446655440000",
  "type": "natural_language",
  "message": {
    "role": "assistant",
    "content": "Hello! I'm here to help you. How can I assist you today?"
  }
}
```

**Response Fields**:
- `conversation_id`: Unique identifier for the conversation
- `type`: Conversation type
- `message`: Assistant's response (if initial_message provided)
  - `role`: Always `assistant`
  - `content`: Response message

**Error Responses**:

`400 Bad Request` - Invalid input:
```json
{
  "error": "Key: 'CreateConversationRequest.Type' Error:Field validation for 'Type' failed on the 'required' tag"
}
```

`500 Internal Server Error` - Database error:
```json
{
  "error": "failed to create conversation"
}
```

**Example Usage**:
```bash
curl -X POST http://localhost:8080/api/v1/conversations \
  -H "Content-Type: application/json" \
  -d '{
    "type": "natural_language",
    "initial_message": "I want to return an item"
  }'
```

---

### Get Conversation

**Endpoint**: `GET /api/v1/conversations/:id`  
**Purpose**: Retrieve conversation details with all messages  
**Authentication**: Customer ID (demo mode in MVP)

**Path Parameters**:
- `id`: Conversation UUID

**Response**: `200 OK`
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "customer_id": "demo-customer-123",
  "type": "natural_language",
  "intent": "return",
  "context": {
    "order_number": "ORD-123456"
  },
  "messages": [
    {
      "id": "msg-uuid-1",
      "role": "user",
      "content": "I want to return order #123456",
      "metadata": {},
      "created_at": "2026-02-04T10:30:00Z"
    },
    {
      "id": "msg-uuid-2",
      "role": "assistant",
      "content": "I can help you with that return.",
      "metadata": {},
      "created_at": "2026-02-04T10:30:01Z"
    }
  ],
  "created_at": "2026-02-04T10:30:00Z",
  "updated_at": "2026-02-04T10:30:01Z",
  "completed_at": null
}
```

**Response Fields**:
- `id`: Conversation UUID
- `customer_id`: Customer identifier
- `type`: `natural_language` or `guided_flow`
- `intent`: Classified intent (if set)
- `context`: Key-value context store
- `messages`: Array of messages (chronological order)
- `created_at`: Conversation start timestamp
- `updated_at`: Last modification timestamp
- `completed_at`: Completion timestamp (null if active)

**Error Responses**:

`404 Not Found` - Conversation doesn't exist:
```json
{
  "error": "conversation not found"
}
```

**Example Usage**:
```bash
curl http://localhost:8080/api/v1/conversations/550e8400-e29b-41d4-a716-446655440000
```

---

### Get Conversation Messages

**Endpoint**: `GET /api/v1/conversations/:id/messages`  
**Purpose**: Retrieve only messages for a conversation  
**Authentication**: Customer ID (demo mode in MVP)

**Path Parameters**:
- `id`: Conversation UUID

**Response**: `200 OK`
```json
{
  "messages": [
    {
      "id": "msg-uuid-1",
      "role": "user",
      "content": "I want to return an item",
      "metadata": {},
      "created_at": "2026-02-04T10:30:00Z"
    },
    {
      "id": "msg-uuid-2",
      "role": "assistant",
      "content": "I can help with your return.",
      "metadata": {},
      "created_at": "2026-02-04T10:30:01Z"
    }
  ]
}
```

**Error Responses**:

`404 Not Found` - Conversation doesn't exist:
```json
{
  "error": "conversation not found"
}
```

**Example Usage**:
```bash
curl http://localhost:8080/api/v1/conversations/550e8400-e29b-41d4-a716-446655440000/messages
```

---

### Send Message

**Endpoint**: `POST /api/v1/conversations/:id/messages`  
**Purpose**: Add a message to an existing conversation  
**Authentication**: Customer ID (demo mode in MVP)

**Path Parameters**:
- `id`: Conversation UUID

**Request Body**:
```json
{
  "content": "I want to return order #123456",
  "metadata": {
    "platform": "mobile_app",
    "version": "2.1.0"
  }
}
```

**Request Fields**:
- `content` (required): Message text
- `metadata` (optional): Additional message metadata

**Response**: `200 OK`
```json
{
  "user_message": {
    "id": "msg-user-uuid",
    "role": "user",
    "content": "I want to return order #123456",
    "metadata": {
      "platform": "mobile_app",
      "version": "2.1.0"
    },
    "created_at": "2026-02-04T10:30:00Z"
  },
  "assistant_message": {
    "id": "msg-assistant-uuid",
    "role": "assistant",
    "content": "Thank you for your message. How can I assist you further?",
    "metadata": {},
    "created_at": "2026-02-04T10:30:01Z"
  }
}
```

**Response Fields**:
- `user_message`: Customer's message (as stored)
- `assistant_message`: System's response

**Error Responses**:

`400 Bad Request` - Invalid input:
```json
{
  "error": "Key: 'SendMessageRequest.Content' Error:Field validation for 'Content' failed on the 'required' tag"
}
```

`404 Not Found` - Conversation doesn't exist:
```json
{
  "error": "conversation not found"
}
```

`500 Internal Server Error` - Conversation already completed:
```json
{
  "error": "conversation is already completed"
}
```

**Example Usage**:
```bash
curl -X POST http://localhost:8080/api/v1/conversations/550e8400-e29b-41d4-a716-446655440000/messages \
  -H "Content-Type: application/json" \
  -d '{
    "content": "I want to return order #123456"
  }'
```

---

## Return Request Endpoints

### Create Return Request

**Endpoint**: `POST /api/v1/returns`  
**Purpose**: Submit a new return request  
**Authentication**: Customer ID (demo mode in MVP)

**Request Body**:
```json
{
  "conversation_id": "550e8400-e29b-41d4-a716-446655440000",
  "order_number": "ORD-123456",
  "items": [
    {
      "order_item_id": "ITEM-1",
      "product_id": "PROD-456",
      "quantity": 1
    }
  ],
  "reason": "wrong_size",
  "detailed_reason": "The shirt is too small for me",
  "photos": ["https://example.com/photo1.jpg"],
  "refund_method": "original_payment",
  "delivery_method": "collect",
  "collection_point": "Store Cape Town CBD"
}
```

**Request Fields**:
- `conversation_id` (optional): Link to conversation
- `order_number` (required): Order identifier
- `items` (required, min 1): Array of items to return
  - `order_item_id` (required): Item ID from order
  - `product_id` (required): Product identifier
  - `quantity` (required, min 1): Quantity to return
- `reason` (required): Return reason
- `detailed_reason` (optional): Additional details
- `photos` (optional): Array of photo URLs
- `refund_method` (required): `original_payment`, `gift_card`, or `bash_account`
- `delivery_method` (required): `collect` or `ship`
- `collection_point` (conditional): Required if delivery_method is `collect`
- `shipping_address` (conditional): Required if delivery_method is `ship`

**Response**: `201 Created`
```json
{
  "id": "return-uuid",
  "return_reference": "RET-abcd1234",
  "status": "approved",
  "estimated_refund_date": "2026-02-11T10:00:00Z",
  "refund_amount_cents": 15999,
  "message": "Your return has been submitted successfully."
}
```

**Response Fields**:
- `id`: Return request UUID
- `return_reference`: Human-readable reference (format: `RET-{prefix}`)
- `status`: Current status (automatically `approved` in MVP)
- `estimated_refund_date`: Expected refund date (7 days from submission)
- `refund_amount_cents`: Refund amount in cents
- `message`: Confirmation message

**Error Responses**:

`400 Bad Request` - Validation failure:
```json
{
  "error": "Key: 'CreateReturnRequest.OrderNumber' Error:Field validation for 'OrderNumber' failed on the 'required' tag"
}
```

`400 Bad Request` - Order not found:
```json
{
  "error": "order not found"
}
```

`400 Bad Request` - Order ownership mismatch:
```json
{
  "error": "order does not belong to customer"
}
```

`400 Bad Request` - Invalid items:
```json
{
  "error": "item ITEM-1 not found in order ORD-123456"
}
```

`500 Internal Server Error` - Database error:
```json
{
  "error": "failed to create return request"
}
```

**Example Usage**:
```bash
curl -X POST http://localhost:8080/api/v1/returns \
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
    "collection_point": "Store Cape Town CBD"
  }'
```

---

### Get Return Request

**Endpoint**: `GET /api/v1/returns/:id`  
**Purpose**: Retrieve return request details  
**Authentication**: Customer ID (demo mode in MVP)

**Path Parameters**:
- `id`: Return request UUID

**Response**: `200 OK`
```json
{
  "id": "return-uuid",
  "conversation_id": "conv-uuid",
  "customer_id": "demo-customer-123",
  "order_number": "ORD-123456",
  "items": [
    {
      "order_item_id": "ITEM-1",
      "product_id": "PROD-456",
      "quantity": 1
    }
  ],
  "reason": "wrong_size",
  "detailed_reason": "The shirt is too small",
  "photos": ["https://example.com/photo1.jpg"],
  "refund_method": "original_payment",
  "delivery_method": "collect",
  "collection_point": "Store Cape Town CBD",
  "status": "approved",
  "estimated_refund_date": "2026-02-11T10:00:00Z",
  "refund_amount_cents": 15999,
  "return_reference": "RET-abcd1234",
  "created_at": "2026-02-04T10:00:00Z",
  "updated_at": "2026-02-04T10:00:00Z"
}
```

**Error Responses**:

`404 Not Found` - Return request doesn't exist:
```json
{
  "error": "return request not found"
}
```

**Example Usage**:
```bash
curl http://localhost:8080/api/v1/returns/return-uuid
```

---

### List Return Requests

**Endpoint**: `GET /api/v1/returns`  
**Purpose**: List all return requests for customer  
**Authentication**: Customer ID (demo mode in MVP)

**Query Parameters**: None (customer ID from auth context)

**Response**: `200 OK`
```json
{
  "returns": [
    {
      "id": "return-uuid-1",
      "order_number": "ORD-123456",
      "status": "approved",
      "refund_amount_cents": 15999,
      "return_reference": "RET-abcd1234",
      "created_at": "2026-02-04T10:00:00Z"
    },
    {
      "id": "return-uuid-2",
      "order_number": "ORD-789012",
      "status": "refunded",
      "refund_amount_cents": 5998,
      "return_reference": "RET-efgh5678",
      "created_at": "2026-01-28T14:30:00Z"
    }
  ]
}
```

**Response Fields**:
- `returns`: Array of return request summaries (chronological order, newest first)

**Error Responses**:

`500 Internal Server Error` - Database error:
```json
{
  "error": "failed to list return requests"
}
```

**Example Usage**:
```bash
curl http://localhost:8080/api/v1/returns
```

---

## Future Endpoints

### Complete Conversation
**Endpoint**: `POST /api/v1/conversations/:id/complete`  
**Purpose**: Mark conversation as completed

### Update Return Status
**Endpoint**: `PATCH /api/v1/returns/:id/status`  
**Purpose**: Update return request status (admin)

### Upload Photos
**Endpoint**: `POST /api/v1/returns/:id/photos`  
**Purpose**: Upload return item photos

### Approve/Reject Return
**Endpoint**: `POST /api/v1/returns/:id/approve`  
**Endpoint**: `POST /api/v1/returns/:id/reject`  
**Purpose**: Manual return approval/rejection

---

## Rate Limiting

**Current**: None (MVP)  
**Future**: Rate limiting based on customer ID or IP address

---

## Pagination

**Current**: No pagination (all results returned)  
**Future**: Query parameters for pagination
- `page`: Page number (default: 1)
- `page_size`: Results per page (default: 20)

---

## API Versioning Strategy

**Current Version**: v1 (in URL path)  
**Breaking Changes**: New major version (v2, v3, etc.)  
**Non-Breaking Changes**: Same version with added fields/endpoints

**Version Support**: Only current version supported in MVP
