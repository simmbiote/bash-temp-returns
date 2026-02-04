# Customer Support Chatbot API - Go Implementation Specification

## Overview

A Go-based REST API that powers a comprehensive customer support chatbot, enabling customers to perform self-service tasks through both natural language and structured guided flows. While the initial focus is on **Returns processing**, the architecture is designed to support multiple customer support flows including:

- **Returns** (Phase 1 - Primary focus)
- **Refunds** (Track refund status, request refund method changes)
- **Authentication/Login** (Password reset, account recovery, OTP verification)
- **Account Management** (Update profile, change address, manage preferences)
- **Order Tracking** (Real-time tracking, delivery updates)
- **Product Inquiries** (Stock availability, product information)
- **General Support** (FAQs, policy questions, escalation to human agents)

The API integrates with the existing Orders domain and is built with an extensible flow configuration system. **Note**: Admin interface for flow management is out of scope for this phase - flows will be configured via database seeding and API endpoints.

## Architecture Overview

### Tech Stack

- **Language**: Go 1.21+
- **Framework**: Gin (HTTP routing)
- **Architecture**: Clean Architecture / Hexagonal Architecture
- **Auth**: JWT validation with middleware
- **AI Integration**: OpenAI/Anthropic SDK for natural language processing (or keyword-based for MVP)
- **Database**: SQLite (local/MVP) or PostgreSQL (production) - conversation history, flow configurations
- **Cache**: In-memory (MVP) or Redis (production) - session state, order data
- **API Client**: HTTP client for Orders API integration

**Note**: This spec supports both SQLite (for local development without Docker) and PostgreSQL (for production). JSONB columns are stored as TEXT in SQLite with JSON marshaling in Go.

### Project Structure

```
customer-support-api/
├── cmd/
│   └── api/
│       └── main.go
├── internal/
│   ├── domain/
│   │   ├── entities/
│   │   │   ├── conversation.go
│   │   │   ├── return_request.go
│   │   │   ├── refund_request.go
│   │   │   ├── account_action.go
│   │   │   ├── flow.go
│   │   │   └── order.go
│   │   ├── repositories/
│   │   │   ├── conversation_repository.go
│   │   │   ├── return_repository.go
│   │   │   ├── refund_repository.go
│   │   │   ├── account_repository.go
│   │   │   └── flow_repository.go
│   │   └── services/
│   │       ├── conversation_service.go
│   │       ├── return_service.go
│   │       ├── refund_service.go
│   │       ├── account_service.go
│   │       └── flow_service.go
│   ├── application/
│   │   ├── usecases/
│   │   │   ├── process_message.go
│   │   │   ├── initiate_return.go
│   │   │   ├── complete_return.go
│   │   │   ├── process_refund_inquiry.go
│   │   │   ├── handle_account_action.go
│   │   │   └── get_order_history.go
│   │   └── dto/
│   │       ├── message_request.go
│   │       ├── flow_action_request.go
│   │       └── return_request.go
│   ├── infrastructure/
│   │   ├── api/
│   │   │   ├── orders_client.go
│   │   │   └── return_management_client.go
│   │   ├── ai/
│   │   │   ├── openai_client.go
│   │   │   └── intent_classifier.go
│   │   ├── persistence/
│   │   │   ├── postgres/
│   │   │   └── redis/
│   │   └── middleware/
│   │       ├── auth.go
│   │       ├── logging.go
│   │       └── rate_limit.go
│   ├── interfaces/
│   │   ├── http/
│   │   │   ├── handlers/
│   │   │   │   ├── conversation_handler.go
│   │   │   │   ├── return_handler.go
│   │   │   │   └── order_handler.go
│   │   │   ├── middleware/
│   │   │   │   └── customer_auth.go
│   │   │   └── router.go
│   │   └── websocket/
│   │       └── chat_handler.go
│   └── pkg/
│       ├── errors/
│       │   └── errors.go
│       ├── logger/
│       │   └── logger.go
│       └── validator/
│           └── validator.go
├── config/
│   ├── config.go
│   └── config.yaml
├── migrations/
│   └── 001_initial_schema.sql
├── tests/
│   ├── integration/
│   └── unit/
├── docs/
│   ├── api.yaml (OpenAPI spec)
│   └── flows.md
├── go.mod
├── go.sum
├── Makefile
└── Dockerfile
```

## Core Domain Entities

### 1. Conversation

```go
package entities

import (
    "time"
    "github.com/google/uuid"
)

type ConversationType string

const (
    ConversationTypeNaturalLanguage ConversationType = "natural_language"
    ConversationTypeGuidedFlow      ConversationType = "guided_flow"
)

type Conversation struct {
    ID          uuid.UUID         `json:"id"`
    CustomerID  string            `json:"customer_id"`
    Type        ConversationType  `json:"type"`
    Intent      string            `json:"intent"` // "return", "track_order", "change_address"
    State       ConversationState `json:"state"`
    Context     map[string]any    `json:"context"`
    Messages    []Message         `json:"messages"`
    CreatedAt   time.Time         `json:"created_at"`
    UpdatedAt   time.Time         `json:"updated_at"`
    CompletedAt *time.Time        `json:"completed_at,omitempty"`
}

type ConversationState struct {
    CurrentStep string         `json:"current_step"`
    FlowID      *uuid.UUID     `json:"flow_id,omitempty"`
    Data        map[string]any `json:"data"`
}

type Message struct {
    ID         uuid.UUID   `json:"id"`
    Role       MessageRole `json:"role"` // "user", "assistant", "system"
    Content    string      `json:"content"`
    Metadata   Metadata    `json:"metadata,omitempty"`
    Timestamp  time.Time   `json:"timestamp"`
}

type MessageRole string

const (
    MessageRoleUser      MessageRole = "user"
    MessageRoleAssistant MessageRole = "assistant"
    MessageRoleSystem    MessageRole = "system"
)

type Metadata struct {
    IntentConfidence float64        `json:"intent_confidence,omitempty"`
    ExtractedData    map[string]any `json:"extracted_data,omitempty"`
    Options          []Option       `json:"options,omitempty"`
}

type Option struct {
    ID    string `json:"id"`
    Label string `json:"label"`
    Value string `json:"value"`
}
```

### 2. Return Request

```go
package entities

import (
    "time"
    "github.com/google/uuid"
)

type ReturnStatus string

const (
    ReturnStatusDraft            ReturnStatus = "draft"
    ReturnStatusPendingApproval  ReturnStatus = "pending_approval"
    ReturnStatusApproved         ReturnStatus = "approved"
    ReturnStatusRejected         ReturnStatus = "rejected"
    ReturnStatusCompleted        ReturnStatus = "completed"
)

type ReturnRequest struct {
    ID                  uuid.UUID      `json:"id"`
    ConversationID      uuid.UUID      `json:"conversation_id"`
    CustomerID          string         `json:"customer_id"`
    OrderNumber         string         `json:"order_number"`
    Items               []ReturnItem   `json:"items"`
    Reason              string         `json:"reason"`
    DetailedReason      string         `json:"detailed_reason,omitempty"`
    Photos              []string       `json:"photos,omitempty"`
    RefundMethod        RefundMethod   `json:"refund_method"`
    DeliveryMethod      DeliveryMethod `json:"delivery_method"`
    CollectionPoint     *CollectionPoint `json:"collection_point,omitempty"`
    ShippingAddress     *Address       `json:"shipping_address,omitempty"`
    Status              ReturnStatus   `json:"status"`
    EstimatedRefundDate *time.Time     `json:"estimated_refund_date,omitempty"`
    RefundAmountCents   int64          `json:"refund_amount_cents"`
    ReturnReference     string         `json:"return_reference,omitempty"`
    CreatedAt           time.Time      `json:"created_at"`
    UpdatedAt           time.Time      `json:"updated_at"`
}

type ReturnItem struct {
    ItemID   string `json:"item_id"`
    Name     string `json:"name"`
    Quantity int    `json:"quantity"`
    Price    int64  `json:"price_cents"`
}

type RefundMethod string

const (
    RefundMethodOriginalPayment RefundMethod = "original_payment"
    RefundMethodStoreCredit     RefundMethod = "store_credit"
    RefundMethodTFGMoney        RefundMethod = "tfg_money"
)

type DeliveryMethod string

const (
    DeliveryMethodCourier        DeliveryMethod = "courier"
    DeliveryMethodCollectionPoint DeliveryMethod = "collection_point"
)

type CollectionPoint struct {
    ID      string  `json:"id"`
    Name    string  `json:"name"`
    Address Address `json:"address"`
}

type Address struct {
    Street     string `json:"street"`
    City       string `json:"city"`
    Province   string `json:"province"`
    PostalCode string `json:"postal_code"`
    Country    string `json:"country"`
}
```

### 3. Refund Request

```go
package entities

import (
    "time"
    "github.com/google/uuid"
)

type RefundStatus string

const (
    RefundStatusPending    RefundStatus = "pending"
    RefundStatusProcessing RefundStatus = "processing"
    RefundStatusCompleted  RefundStatus = "completed"
    RefundStatusFailed     RefundStatus = "failed"
)

type RefundRequest struct {
    ID                uuid.UUID    `json:"id"`
    ConversationID    uuid.UUID    `json:"conversation_id,omitempty"`
    CustomerID        string       `json:"customer_id"`
    OrderNumber       string       `json:"order_number"`
    ReturnReference   string       `json:"return_reference,omitempty"`
    RefundAmountCents int64        `json:"refund_amount_cents"`
    RefundMethod      RefundMethod `json:"refund_method"`
    Status            RefundStatus `json:"status"`
    ProcessedDate     *time.Time   `json:"processed_date,omitempty"`
    FailureReason     string       `json:"failure_reason,omitempty"`
    CreatedAt         time.Time    `json:"created_at"`
    UpdatedAt         time.Time    `json:"updated_at"`
}
```

### 4. Account Action

```go
package entities

import (
    "time"
    "github.com/google/uuid"
)

type AccountActionType string

const (
    AccountActionPasswordReset   AccountActionType = "password_reset"
    AccountActionUpdateEmail     AccountActionType = "update_email"
    AccountActionUpdatePhone     AccountActionType = "update_phone"
    AccountActionUpdateAddress   AccountActionType = "update_address"
    AccountActionDeleteAccount   AccountActionType = "delete_account"
    AccountActionOTPVerification AccountActionType = "otp_verification"
)

type AccountActionStatus string

const (
    AccountActionStatusPending   AccountActionStatus = "pending"
    AccountActionStatusVerifying AccountActionStatus = "verifying"
    AccountActionStatusCompleted AccountActionStatus = "completed"
    AccountActionStatusFailed    AccountActionStatus = "failed"
    AccountActionStatusCancelled AccountActionStatus = "cancelled"
)

type AccountAction struct {
    ID             uuid.UUID           `json:"id"`
    ConversationID uuid.UUID           `json:"conversation_id,omitempty"`
    CustomerID     string              `json:"customer_id"`
    ActionType     AccountActionType   `json:"action_type"`
    Status         AccountActionStatus `json:"status"`
    OTPCode        string              `json:"otp_code,omitempty"`
    OTPExpiresAt   *time.Time          `json:"otp_expires_at,omitempty"`
    OTPAttempts    int                 `json:"otp_attempts"`
    ActionData     map[string]any      `json:"action_data"` // Flexible data for different action types
    VerifiedAt     *time.Time          `json:"verified_at,omitempty"`
    CompletedAt    *time.Time          `json:"completed_at,omitempty"`
    FailureReason  string              `json:"failure_reason,omitempty"`
    CreatedAt      time.Time           `json:"created_at"`
    UpdatedAt      time.Time           `json:"updated_at"`
}
```

### 5. Flow Configuration

```go
package entities

import (
    "time"
    "github.com/google/uuid"
)

type Flow struct {
    ID          uuid.UUID  `json:"id"`
    Name        string     `json:"name"`
    Intent      string     `json:"intent"`
    Version     int        `json:"version"`
    Steps       []FlowStep `json:"steps"`
    IsActive    bool       `json:"is_active"`
    CreatedAt   time.Time  `json:"created_at"`
    UpdatedAt   time.Time  `json:"updated_at"`
}

type FlowStep struct {
    ID          string            `json:"id"`
    Type        StepType          `json:"type"`
    Prompt      string            `json:"prompt"`
    Options     []Option          `json:"options,omitempty"`
    Validation  *ValidationRule   `json:"validation,omitempty"`
    NextSteps   map[string]string `json:"next_steps"` // condition -> next step ID
    IsTerminal  bool              `json:"is_terminal"`
}

type StepType string

const (
    StepTypeMessage      StepType = "message"
    StepTypeSingleChoice StepType = "single_choice"
    StepTypeMultiChoice  StepType = "multi_choice"
    StepTypeTextInput    StepType = "text_input"
    StepTypeFileUpload   StepType = "file_upload"
    StepTypeConfirmation StepType = "confirmation"
)

type ValidationRule struct {
    Type     string `json:"type"` // "required", "regex", "min_length", "max_length"
    Value    string `json:"value,omitempty"`
    Message  string `json:"message"`
}
```

## Supported Intents & Flow Patterns

The chatbot supports multiple customer service intents, each with its own flow pattern. Flows can be configured statically (via database seeds/migrations) or dynamically (via internal API endpoints). The system intelligently routes conversations based on detected intent.

### Intent Classification

Intents are classified using AI (OpenAI/Anthropic) for natural language conversations, or selected explicitly in guided flows. Confidence scores determine whether to ask for clarification.

#### Primary Intents (Phase 1)

| Intent | Description | Confidence Threshold | Priority |
|--------|-------------|---------------------|----------|
| `return` | Customer wants to return purchased items | 0.80 | High |
| `track_order` | Customer wants to track order delivery | 0.85 | High |
| `general_inquiry` | General questions, fallback intent | 0.60 | Low |

#### Future Intents (Phase 2+)

| Intent | Description | Triggers |
|--------|-------------|----------|
| `refund_status` | Check refund processing status | "refund", "money back", "reimbursement" |
| `change_address` | Update delivery address | "change address", "wrong address", "different location" |
| `cancel_order` | Cancel pending order | "cancel order", "don't want anymore" |
| `password_reset` | Reset account password | "forgot password", "reset password", "can't login" |
| `update_profile` | Update account information | "change email", "update phone", "edit details" |
| `product_inquiry` | Product availability/info | "in stock", "product details", "sizes available" |
| `complaint` | Lodge complaint, escalate issue | "complaint", "manager", "not satisfied", "speak to someone" |

### Flow Patterns

#### 1. Returns Flow (Implemented - Phase 1)

**Steps:**
1. **Identify Order** - Find order by number or recent orders
2. **Select Items** - Choose which items to return
3. **Specify Reason** - Select return reason + optional details
4. **Upload Photos** - Optional photo evidence
5. **Choose Refund Method** - Original payment, store credit, or TFG Money
6. **Select Delivery Method** - Courier pickup or drop-off at collection point
7. **Enter Address/Collection Point** - Based on delivery method
8. **Confirmation** - Review and confirm return request

**Natural Language Example:**
```
User: "I want to return my Nike shoes. They're too small."
Bot: "I can help you return your Nike shoes. Let me look up your recent orders."
     [Searches orders, finds Nike shoes]
Bot: "I found Order #12345 with Nike Air Max 270 (Size 9). Is this correct?"
User: "Yes"
Bot: "Got it. What's the main reason for returning these shoes?"
     [Presents options: Wrong Size, Quality Issue, Changed Mind, Other]
```

#### 2. Refund Status Flow (Future - Phase 2)

**Steps:**
1. **Identify Return** - By return reference or order number
2. **Fetch Status** - Query refund processing status
3. **Display Timeline** - Show refund progress with estimated dates
4. **Offer Actions** - Change refund method if still processing

**Natural Language Example:**
```
User: "Where's my refund for order #12345?"
Bot: "Let me check the refund status for order #12345..."
Bot: "Your refund of R1,299.00 is currently being processed. It was initiated on Feb 1 and should reach your account by Feb 8. Would you like me to send you tracking updates via email?"
```

#### 3. Password Reset Flow (Future - Phase 2)

**Steps:**
1. **Verify Identity** - Request email or phone
2. **Send OTP** - Generate and send verification code
3. **Verify OTP** - Validate code (max 3 attempts)
4. **Create New Password** - Prompt for new password with strength validation
5. **Confirmation** - Password updated successfully

**Guided Flow Example:**
```
User: [Selects "Reset Password" from menu]
Bot: "I'll help you reset your password. Please enter your email address."
User: "customer@example.com"
Bot: "I've sent a 6-digit verification code to customer@example.com. Please enter it here. (Valid for 10 minutes)"
User: "123456"
Bot: "Code verified! Now create a new password (min 8 characters, must include letters and numbers)."
```

#### 4. Address Update Flow (Future - Phase 2)

**Steps:**
1. **Identify Order** (if changing for active order)
2. **Display Current Address**
3. **Capture New Address** - Street, city, province, postal code
4. **Validate Address** - Check deliverability
5. **Confirmation** - Update address

**Natural Language Example:**
```
User: "I need to change my delivery address for order #12345"
Bot: "I can help you update the delivery address. Your order is currently being delivered to: 123 Old St, Cape Town. What's the new address?"
User: "456 New Street, Johannesburg, 2000"
Bot: "Got it. I've updated your delivery address to 456 New Street, Johannesburg, 2000. The courier has been notified."
```

#### 5. Account Management Flow (Future - Phase 2)

**Steps:**
1. **Select Action Type** - Update email, phone, or preferences
2. **Verify Identity** - Send OTP to current contact
3. **Capture New Information**
4. **Verify New Contact** - Send OTP to new email/phone
5. **Confirmation** - Account updated

#### 6. Order Tracking Flow (Future - Phase 2)

**Steps:**
1. **Identify Order** - By order number or select from recent
2. **Fetch Tracking Info** - Get parcel tracking data
3. **Display Status** - Show delivery timeline and current location
4. **Offer Actions** - Contact courier, report issue, change address

**Natural Language Example:**
```
User: "Where is my order?"
Bot: "Let me check your recent orders..."
Bot: "Your order #12345 is out for delivery! It's currently 5km away and should arrive by 3pm today. Would you like real-time tracking updates via SMS?"
```

### Flow Configuration Structure

Flows are stored as JSON in the database. Here's an example Returns flow configuration:

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "Standard Returns Flow",
  "intent": "return",
  "version": 1,
  "is_active": true,
  "steps": [
    {
      "id": "identify_order",
      "type": "single_choice",
      "prompt": "Which order contains the item you'd like to return?",
      "options": [], // Dynamically populated from customer orders
      "validation": {
        "type": "required",
        "message": "Please select an order"
      },
      "next_steps": {
        "default": "select_items"
      }
    },
    {
      "id": "select_items",
      "type": "multi_choice",
      "prompt": "Which item(s) would you like to return?",
      "options": [], // Dynamically populated from order items
      "validation": {
        "type": "required",
        "message": "Please select at least one item"
      },
      "next_steps": {
        "default": "select_reason"
      }
    },
    {
      "id": "select_reason",
      "type": "single_choice",
      "prompt": "What's the reason for the return?",
      "options": [
        {"id": "wrong_size", "label": "Wrong size", "value": "wrong_size"},
        {"id": "quality", "label": "Quality issue", "value": "quality_issue"},
        {"id": "changed_mind", "label": "Changed my mind", "value": "changed_mind"},
        {"id": "other", "label": "Other", "value": "other"}
      ],
      "next_steps": {
        "other": "detailed_reason",
        "default": "upload_photos"
      }
    },
    {
      "id": "detailed_reason",
      "type": "text_input",
      "prompt": "Please provide more details about your return reason",
      "validation": {
        "type": "min_length",
        "value": "10",
        "message": "Please provide at least 10 characters"
      },
      "next_steps": {
        "default": "upload_photos"
      }
    },
    {
      "id": "upload_photos",
      "type": "file_upload",
      "prompt": "Would you like to upload photos of the item? (Optional)",
      "validation": null,
      "next_steps": {
        "default": "select_refund_method"
      }
    },
    {
      "id": "select_refund_method",
      "type": "single_choice",
      "prompt": "How would you like to receive your refund?",
      "options": [
        {"id": "original", "label": "Original payment method", "value": "original_payment"},
        {"id": "credit", "label": "Store credit", "value": "store_credit"},
        {"id": "tfg", "label": "TFG Money", "value": "tfg_money"}
      ],
      "next_steps": {
        "default": "select_delivery"
      }
    },
    {
      "id": "select_delivery",
      "type": "single_choice",
      "prompt": "How will you return the item?",
      "options": [
        {"id": "courier", "label": "Courier pickup", "value": "courier"},
        {"id": "dropoff", "label": "Drop off at collection point", "value": "collection_point"}
      ],
      "next_steps": {
        "courier": "enter_address",
        "collection_point": "select_collection_point"
      }
    },
    {
      "id": "enter_address",
      "type": "text_input",
      "prompt": "Please enter your pickup address",
      "validation": {
        "type": "required",
        "message": "Address is required"
      },
      "next_steps": {
        "default": "confirmation"
      }
    },
    {
      "id": "select_collection_point",
      "type": "single_choice",
      "prompt": "Select a collection point near you",
      "options": [], // Dynamically populated based on location
      "next_steps": {
        "default": "confirmation"
      }
    },
    {
      "id": "confirmation",
      "type": "confirmation",
      "prompt": "Please review your return request",
      "is_terminal": true,
      "next_steps": {}
    }
  ]
}
```

## API Endpoints

### 1. Conversation Endpoints

#### POST /api/v1/conversations

Start a new conversation (natural language or guided flow).

**Request**:
```json
{
  "type": "natural_language",
  "initial_message": "I want to return my Nike shoes. They're too small."
}
```

**Response**:
```json
{
  "conversation_id": "550e8400-e29b-41d4-a716-446655440000",
  "type": "natural_language",
  "intent": "return",
  "message": {
    "role": "assistant",
    "content": "I can help you return your Nike shoes. Let me look up your recent orders.",
    "metadata": {
      "intent_confidence": 0.95,
      "extracted_data": {
        "product_mention": "Nike shoes",
        "reason_mention": "too small"
      }
    }
  }
}
```

#### POST /api/v1/conversations/:id/messages

Send a message in an existing conversation.

**Request**:
```json
{
  "content": "Yes, order #12345"
}
```

**Response**:
```json
{
  "message": {
    "role": "assistant",
    "content": "I found your order #12345 with the following items:\n\n1. Nike Air Max 270 - Size 9 - R1,299.00\n2. Nike Dri-FIT T-Shirt - Size L - R399.00\n\nWhich item would you like to return?",
    "metadata": {
      "options": [
        {
          "id": "item_1",
          "label": "Nike Air Max 270 - Size 9",
          "value": "ITEM-001"
        },
        {
          "id": "item_2",
          "label": "Nike Dri-FIT T-Shirt - Size L",
          "value": "ITEM-002"
        }
      ]
    }
  },
  "state": {
    "current_step": "select_items",
    "data": {
      "order_number": "12345",
      "order_items": [...]
    }
  }
}
```

#### POST /api/v1/conversations/:id/flow-action

Execute a flow action (for guided flows).

**Request**:
```json
{
  "action": "select_option",
  "value": "ITEM-001"
}
```

**Response**:
```json
{
  "message": {
    "role": "assistant",
    "content": "You've selected Nike Air Max 270. What's the reason for the return?",
    "metadata": {
      "options": [
        {"id": "size", "label": "Wrong size", "value": "wrong_size"},
        {"id": "quality", "label": "Quality issue", "value": "quality_issue"},
        {"id": "changed_mind", "label": "Changed my mind", "value": "changed_mind"},
        {"id": "other", "label": "Other", "value": "other"}
      ]
    }
  },
  "state": {
    "current_step": "select_reason",
    "data": {
      "selected_items": ["ITEM-001"]
    }
  }
}
```

#### GET /api/v1/conversations/:id

Retrieve conversation history.

**Response**:
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "type": "natural_language",
  "intent": "return",
  "state": {
    "current_step": "select_items",
    "data": {...}
  },
  "messages": [
    {
      "role": "user",
      "content": "I want to return my Nike shoes",
      "timestamp": "2026-02-04T10:30:00Z"
    },
    {
      "role": "assistant",
      "content": "I can help you return your Nike shoes...",
      "timestamp": "2026-02-04T10:30:02Z"
    }
  ],
  "created_at": "2026-02-04T10:30:00Z",
  "updated_at": "2026-02-04T10:35:00Z"
}
```

### 2. Return Endpoints

#### POST /api/v1/returns

Create a return request (can be called from conversation or directly).

**Request**:
```json
{
  "conversation_id": "550e8400-e29b-41d4-a716-446655440000",
  "order_number": "12345",
  "items": [
    {
      "item_id": "ITEM-001",
      "quantity": 1
    }
  ],
  "reason": "wrong_size",
  "detailed_reason": "Shoes are too small, need a size up",
  "photos": [
    "https://storage.example.com/returns/photo1.jpg"
  ],
  "refund_method": "original_payment",
  "delivery_method": "courier",
  "shipping_address": {
    "street": "123 Main Street",
    "city": "Cape Town",
    "province": "Western Cape",
    "postal_code": "8001",
    "country": "ZA"
  }
}
```

**Response**:
```json
{
  "id": "660e8400-e29b-41d4-a716-446655440001",
  "return_reference": "RET-2026-0001",
  "status": "pending_approval",
  "estimated_refund_date": "2026-02-14T00:00:00Z",
  "refund_amount_cents": 129900,
  "message": "Your return has been submitted successfully. You'll receive a confirmation email shortly with return shipping instructions."
}
```

#### GET /api/v1/returns

Get customer's return history.

**Response**:
```json
{
  "returns": [
    {
      "id": "660e8400-e29b-41d4-a716-446655440001",
      "return_reference": "RET-2026-0001",
      "order_number": "12345",
      "status": "pending_approval",
      "created_at": "2026-02-04T10:40:00Z"
    }
  ],
  "total": 1
}
```

#### GET /api/v1/returns/:id

Get detailed return information.

### 3. Refund Endpoints (Future - Phase 2)

#### GET /api/v1/refunds

Get customer's refund history.

**Query Parameters**:
- `order_number` (optional) - Filter by order
- `status` (optional) - Filter by status
- `page` (default: 0)
- `page_size` (default: 20)

**Response**:
```json
{
  "refunds": [
    {
      "id": "770e8400-e29b-41d4-a716-446655440002",
      "order_number": "12345",
      "return_reference": "RET-2026-0001",
      "refund_amount_cents": 129900,
      "refund_method": "original_payment",
      "status": "processing",
      "processed_date": null,
      "created_at": "2026-02-04T10:40:00Z"
    }
  ],
  "total": 1
}
```

#### GET /api/v1/refunds/:id

Get detailed refund information including processing timeline.

**Response**:
```json
{
  "id": "770e8400-e29b-41d4-a716-446655440002",
  "order_number": "12345",
  "return_reference": "RET-2026-0001",
  "refund_amount_cents": 129900,
  "refund_method": "original_payment",
  "status": "processing",
  "timeline": [
    {
      "status": "pending",
      "timestamp": "2026-02-04T10:40:00Z",
      "description": "Refund request received"
    },
    {
      "status": "processing",
      "timestamp": "2026-02-05T09:00:00Z",
      "description": "Refund being processed by finance team"
    }
  ],
  "estimated_completion_date": "2026-02-08T00:00:00Z"
}
```

### 4. Account Management Endpoints (Future - Phase 2)

#### POST /api/v1/account/password-reset

Initiate password reset flow.

**Request**:
```json
{
  "email": "customer@example.com"
}
```

**Response**:
```json
{
  "action_id": "880e8400-e29b-41d4-a716-446655440003",
  "status": "pending",
  "message": "Verification code sent to your email",
  "otp_expires_at": "2026-02-04T11:00:00Z"
}
```

#### POST /api/v1/account/verify-otp

Verify OTP code for account actions.

**Request**:
```json
{
  "action_id": "880e8400-e29b-41d4-a716-446655440003",
  "otp_code": "123456"
}
```

**Response**:
```json
{
  "verified": true,
  "message": "OTP verified successfully",
  "next_step": "set_new_password"
}
```

#### PUT /api/v1/account/profile

Update customer profile information.

**Request**:
```json
{
  "action_type": "update_phone",
  "phone": "+27821234567",
  "verification_required": true
}
```

### 5. Order Endpoints (Proxy to Orders API)

These endpoints proxy requests to the existing Orders BFF API. The Orders API returns responses wrapped in a standard envelope with metadata.

#### GET /api/v1/orders

Get customer's order history.

**Query Parameters**:
- `page` (default: 0)
- `page_size` (default: 20)

**Response** (simplified, actual Orders API response):
```json
{
  "data": {
    "pages": 1,
    "total": 17,
    "page": 0,
    "pageSize": 20,
    "orders": [
      {
        "orderDate": "2025-11-12T20:19:29.867",
        "orderStatus": "Completed",
        "orderNumber": "B22881153-01",
        "orderStatusDescription": "Complete",
        "total": 7475,
        "itemImageUrls": [
          "https://assets.bash.com/products/1000x1000/57206641.jpg"
        ],
        "numberDeliveries": 1,
        "numberPendingItems": 0,
        "shippingMethod": {
          "shippingMethodType": "CollectInStore",
          "shippingMethodName": "Store Collection"
        }
      }
    ]
  },
  "success": true,
  "errorCode": null,
  "errorMessage": null,
  "timestamp": "2026-02-04T11:04:07.036Z",
  "path": "/v1/order/orders/history",
  "request_id": "321a21c3-3339-476a-bbde-bf3ed22c0ca1"
}
```

**Note**: Amounts are in cents (e.g., `7475` = R74.75). The Orders API uses format `B{number}-01` for order numbers.

#### GET /api/v1/orders/:orderNumber

Get detailed order information including parcels, items, and tracking.

**Example**: `/api/v1/orders/B22881153-01`

**Response** (actual Orders API structure):
```json
{
  "data": {
    "orderNumber": "B22881153-01",
    "orderDate": "2025-11-12T20:19:29.867",
    "total": 7475,
    "orderStatus": "Completed",
    "orderStatusDescription": "Complete",
    "parcels": [
      {
        "parcelName": "Collection #1",
        "parcelNumber": 1,
        "parcelStatus": "Delivered",
        "parcelStatusHelperText": null,
        "parcelStatusDescription": "Delivered",
        "parcelInvoices": [
          {
            "url": "https://docs.bash.com/v1/invoice/40fe2dc5-62ee-4a18-998b-e3281c02086c",
            "date": "2025-11-12T22:24:30.0666667"
          }
        ],
        "orderItems": [
          {
            "slug": "usn-energy-oats-bar-double-chocolate-35g-130602aafa5",
            "name": "USN ENERGY OATS BAR DOUBLE CHOCOLATE 35G",
            "size": "",
            "colour": "None",
            "quantity": 1,
            "image": "https://assets.bash.com/products/1000x1000/57206641.jpg",
            "price": 1495,
            "store": "TOTALSPORTS",
            "brand": "Usn",
            "sku": "57206641",
            "soldBy": null
          }
        ],
        "trackingNumber": "CAZA0747786",
        "trackingUrl": "https://picup.co.za/o/YwBEZD",
        "estimatedArrivalDate": null,
        "courier": "bashDelivery"
      }
    ],
    "invoices": [
      {
        "url": "https://docs.bash.com/v1/invoice/40fe2dc5-62ee-4a18-998b-e3281c02086c",
        "date": "2025-11-12T22:24:30.0666667"
      }
    ],
    "subtotal": 7475,
    "discount": 0,
    "shippingCost": 0,
    "deliveryDescription": null,
    "fulfilledIn": "Up to 5 days",
    "itemImageUrls": [
      "https://assets.bash.com/products/1000x1000/57206641.jpg"
    ],
    "numberDeliveries": 1,
    "numberPendingItems": 0,
    "shippingMethod": {
      "shippingMethodType": "CollectInStore",
      "shippingMethodName": "Store Collection"
    },
    "shippingOption": {
      "shippingOptionType": "Standard",
      "shippingOptionName": "Standard"
    },
    "customerDetails": {
      "firstname": "John",
      "surname": "Simms",
      "email": "johnsi@tfg.co.za",
      "mobile": "+27722883520"
    },
    "shippingAddress": {
      "line1": "Sportscene Kenilworth Doncaster Rd",
      "line2": "Shop 65, Kenilworth Centre",
      "line3": "Kenilworth",
      "city": "Cape Town",
      "postalCode": "7708",
      "name": "John Simms",
      "surname": "",
      "mobileNumber": "+27722883520",
      "email": "johnsi@tfg.co.za"
    },
    "isEligibleForReturn": false
  },
  "success": true,
  "errorCode": null,
  "errorMessage": null,
  "timestamp": "2026-02-04T11:06:52.846Z",
  "path": "/v1/order/B22881153-01",
  "request_id": "509dd589-52b1-4247-85b5-477a00223b81"
}
```

**Key Features**:
- **Parcels**: Orders can have multiple parcels with separate tracking
- **Items per Parcel**: Items are organized by parcel, not at order level
- **Tracking**: Each parcel has `trackingNumber`, `trackingUrl`, and `courier` info
- **Invoices**: Both order-level and parcel-level invoices available
- **Addresses**: Collection orders show store address in `shippingAddress`
- **Return Eligibility**: `isEligibleForReturn` flag indicates if order can be returned

#### GET /api/v1/orders/:orderNumber/track

Get order tracking information (extracted from parcel data).

**Response**:
```json
{
  "order_number": "B22881153-01",
  "status": "out_for_delivery",
  "parcels": [
    {
      "parcel_id": "PARCEL-001",
      "tracking_number": "TRK123456789",
      "carrier": "CourierGuy",
      "status": "out_for_delivery",
      "estimated_delivery": "2026-02-04T15:00:00Z",
      "current_location": "Cape Town Distribution Center",
      "tracking_events": [
        {
          "timestamp": "2026-02-03T08:00:00Z",
          "status": "dispatched",
          "location": "Johannesburg Warehouse"
        },
        {
          "timestamp": "2026-02-04T06:00:00Z",
          "status": "in_transit",
          "location": "Cape Town Distribution Center"
        },
        {
          "timestamp": "2026-02-04T10:00:00Z",
          "status": "out_for_delivery",
          "location": "Cape Town Distribution Center"
        }
      ]
    }
  ]
}
```

### 6. Flow Management Endpoints (Internal API - No Admin UI)

These endpoints allow flow configuration via API calls. They should be secured with internal API keys and not exposed to customers.

#### GET /api/v1/internal/flows

List all flow configurations.

**Headers**:
- `X-Internal-API-Key: <secret_key>`

**Query Parameters**:
- `intent` (optional) - Filter by intent
- `is_active` (optional) - Filter active/inactive flows

**Response**:
```json
{
  "flows": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "name": "Standard Returns Flow",
      "intent": "return",
      "version": 1,
      "is_active": true,
      "created_at": "2026-01-15T10:00:00Z",
      "updated_at": "2026-01-15T10:00:00Z"
    }
  ],
  "total": 1
}
```

#### GET /api/v1/internal/flows/:id

Get detailed flow configuration including all steps.

#### POST /api/v1/internal/flows

Create a new flow configuration.

**Request**:
```json
{
  "name": "Express Returns Flow",
  "intent": "return",
  "version": 1,
  "steps": [
    {
      "id": "identify_order",
      "type": "single_choice",
      "prompt": "Which order contains the item you'd like to return?",
      "next_steps": {"default": "select_items"}
    }
  ]
}
```

#### PUT /api/v1/internal/flows/:id

Update an existing flow configuration.

#### POST /api/v1/internal/flows/:id/activate

Activate a specific flow version (deactivates other versions for the same intent).

#### DELETE /api/v1/internal/flows/:id

Soft delete a flow configuration (sets is_active to false).

## Use Cases Implementation

### 1. Process Natural Language Message

```go
package usecases

import (
    "context"
    "github.com/customer-support-api/internal/domain/entities"
    "github.com/customer-support-api/internal/domain/services"
)

type ProcessMessageUseCase struct {
    conversationService *services.ConversationService
    intentClassifier    IntentClassifier
    orderService        OrderService
    returnService       *services.ReturnService
}

func (uc *ProcessMessageUseCase) Execute(
    ctx context.Context,
    conversationID string,
    userMessage string,
) (*ProcessMessageResponse, error) {
    // 1. Load conversation
    conversation, err := uc.conversationService.GetByID(ctx, conversationID)
    if err != nil {
        return nil, err
    }

    // 2. Classify intent and extract entities
    intent, entities, confidence := uc.intentClassifier.Classify(
        ctx,
        userMessage,
        conversation.Messages,
    )

    // 3. Update conversation state
    conversation.Intent = intent
    conversation.Messages = append(conversation.Messages, entities.Message{
        Role:    entities.MessageRoleUser,
        Content: userMessage,
        Metadata: entities.Metadata{
            IntentConfidence: confidence,
            ExtractedData:    entities,
        },
    })

    // 4. Execute intent-specific logic
    var response string
    var options []entities.Option

    switch intent {
    case "return":
        response, options, err = uc.handleReturnIntent(ctx, conversation, entities)
    case "track_order":
        response, options, err = uc.handleTrackOrderIntent(ctx, conversation, entities)
    case "refund_status":
        // Phase 2 intent - provide coming soon message
        response = "Refund tracking is coming soon! For now, please check your email for refund updates or contact our support team."
        options = nil
    case "change_address":
        // Phase 2 intent
        response = "Address change functionality is coming soon! Please contact our support team for urgent address changes."
        options = nil
    case "password_reset":
        // Phase 2 intent
        response = "Password reset via chat is coming soon! For now, please use the 'Forgot Password' link on the login page."
        options = nil
    case "general_inquiry":
        response, options, err = uc.handleGeneralInquiry(ctx, conversation, entities)
    default:
        response = "I'm not sure I understand. Could you please rephrase or choose from these options:"
        options = []entities.Option{
            {ID: "return", Label: "Return an item", Value: "return"},
            {ID: "track", Label: "Track my order", Value: "track_order"},
            {ID: "other", Label: "Something else", Value: "general_inquiry"},
        }
    }

    if err != nil {
        return nil, err
    }

    // 5. Add assistant response
    assistantMessage := entities.Message{
        Role:    entities.MessageRoleAssistant,
        Content: response,
        Metadata: entities.Metadata{
            Options: options,
        },
    }
    conversation.Messages = append(conversation.Messages, assistantMessage)

    // 6. Save conversation
    if err := uc.conversationService.Update(ctx, conversation); err != nil {
        return nil, err
    }

    return &ProcessMessageResponse{
        Message: assistantMessage,
        State:   conversation.State,
    }, nil
}

func (uc *ProcessMessageUseCase) handleReturnIntent(
    ctx context.Context,
    conversation *entities.Conversation,
    extractedData map[string]any,
) (string, []entities.Option, error) {
    // Determine current step in return flow
    currentStep := conversation.State.CurrentStep
    if currentStep == "" {
        currentStep = "identify_order"
    }

    switch currentStep {
    case "identify_order":
        return uc.identifyOrder(ctx, conversation, extractedData)
    case "select_items":
        return uc.selectItems(ctx, conversation, extractedData)
    case "select_reason":
        return uc.selectReason(ctx, conversation)
    case "upload_photos":
        return uc.promptForPhotos(ctx, conversation)
    case "select_delivery":
        return uc.selectDeliveryMethod(ctx, conversation)
    case "confirm":
        return uc.confirmReturn(ctx, conversation)
    default:
        return "Something went wrong. Let's start over.", nil, nil
    }
}

func (uc *ProcessMessageUseCase) identifyOrder(
    ctx context.Context,
    conversation *entities.Conversation,
    extractedData map[string]any,
) (string, []entities.Option, error) {
    // Check if order number was mentioned
    if orderNumber, ok := extractedData["order_number"].(string); ok {
        // Validate order and check eligibility
        order, err := uc.orderService.GetDetailedOrder(ctx, orderNumber)
        if err != nil {
            return "I couldn't find that order. Could you please check the order number?", nil, nil
        }

        if !order.IsEligibleForReturn {
            return "I'm sorry, but this order is not eligible for returns. The return window may have expired.", nil, nil
        }

        // Store order in conversation state
        conversation.State.Data["order"] = order
        conversation.State.CurrentStep = "select_items"

        // Build options from order items (items are nested in parcels)
        var allItems []entities.Option
        for _, parcel := range order.Parcels {
            for _, item := range parcel.OrderItems {
                // Create unique ID combining SKU and a counter if needed
                itemOption := entities.Option{
                    ID:    item.SKU,
                    Label: fmt.Sprintf("%s - R%.2f", item.Name, float64(item.Price)/100),
                    Value: item.SKU,
                }
                allItems = append(allItems, itemOption)
            }
        }

        return fmt.Sprintf("I found your order #%s. Which item(s) would you like to return?", orderNumber), allItems, nil
    }

    // Order not identified, ask for recent orders
    orders, err := uc.orderService.GetOrderHistory(ctx, conversation.CustomerID, 0, 5)
    if err != nil {
        return "I'm having trouble loading your orders. Please try again.", nil, nil
    }

    if len(orders.Orders) == 0 {
        return "I couldn't find any recent orders. Please provide your order number.", nil, nil
    }

    options := make([]entities.Option, len(orders.Orders))
    for i, order := range orders.Orders {
        // Format date nicely
        orderDate, _ := time.Parse(time.RFC3339, order.OrderDate)
        dateStr := orderDate.Format("Jan 2, 2006")
        
        options[i] = entities.Option{
            ID:    order.OrderNumber,
            Label: fmt.Sprintf("Order #%s (%s) - R%.2f", order.OrderNumber, dateStr, float64(order.Total)/100),
            Value: order.OrderNumber,
        }
    }

    return "Here are your recent orders. Which one contains the item you'd like to return?", options, nil
}
```

### 2. Initiate Return (Guided Flow)

```go
package usecases

type InitiateReturnFlowUseCase struct {
    conversationService *services.ConversationService
    flowService         *services.FlowService
}

func (uc *InitiateReturnFlowUseCase) Execute(
    ctx context.Context,
    customerID string,
) (*InitiateReturnFlowResponse, error) {
    // 1. Load active return flow
    flow, err := uc.flowService.GetActiveFlowByIntent(ctx, "return")
    if err != nil {
        return nil, err
    }

    // 2. Create new conversation
    conversation := &entities.Conversation{
        CustomerID: customerID,
        Type:       entities.ConversationTypeGuidedFlow,
        Intent:     "return",
        State: entities.ConversationState{
            FlowID:      &flow.ID,
            CurrentStep: flow.Steps[0].ID,
            Data:        make(map[string]any),
        },
    }

    if err := uc.conversationService.Create(ctx, conversation); err != nil {
        return nil, err
    }

    // 3. Get first step
    firstStep := flow.Steps[0]

    return &InitiateReturnFlowResponse{
        ConversationID: conversation.ID,
        Message: entities.Message{
            Role:    entities.MessageRoleAssistant,
            Content: firstStep.Prompt,
            Metadata: entities.Metadata{
                Options: firstStep.Options,
            },
        },
        FlowStep: firstStep,
    }, nil
}
```

### 3. Handle Track Order Intent

```go
func (uc *ProcessMessageUseCase) handleTrackOrderIntent(
    ctx context.Context,
    conversation *entities.Conversation,
    extractedData map[string]any,
) (string, []entities.Option, error) {
    // Check if order number was extracted
    if orderNumber, ok := extractedData["order_number"].(string); ok {
        // Get tracking information
        tracking, err := uc.orderService.GetParcelTracking(ctx, orderNumber)
        if err != nil {
            return "I couldn't find tracking information for that order. Could you verify the order number?", nil, nil
        }

        // Format tracking response
        response := fmt.Sprintf("Your order #%s is currently: %s\n\n", 
            orderNumber, tracking.Status)
        
        if tracking.EstimatedDelivery != "" {
            response += fmt.Sprintf("Estimated delivery: %s\n", tracking.EstimatedDelivery)
        }
        
        if tracking.CurrentLocation != "" {
            response += fmt.Sprintf("Current location: %s\n", tracking.CurrentLocation)
        }

        response += "\nWould you like real-time updates via SMS?"
        
        options := []entities.Option{
            {ID: "yes_sms", Label: "Yes, send me SMS updates", Value: "enable_sms"},
            {ID: "no_sms", Label: "No, thank you", Value: "no_action"},
        }

        return response, options, nil
    }

    // Order number not extracted - ask for recent orders
    orders, err := uc.orderService.GetOrderHistory(ctx, conversation.CustomerID, 0, 5)
    if err != nil {
        return "I'm having trouble loading your orders. Please provide your order number.", nil, nil
    }

    if len(orders.Orders) == 0 {
        return "I couldn't find any recent orders. Please provide your order number.", nil, nil
    }

    // Show recent orders for selection
    options := make([]entities.Option, len(orders.Orders))
    for i, order := range orders.Orders {
        // Calculate number of items from images array length
        numItems := len(order.ItemImageUrls)
        orderDate, _ := time.Parse(time.RFC3339, order.OrderDate)
        dateStr := orderDate.Format("Jan 2")
        
        options[i] = entities.Option{
            ID:    order.OrderNumber,
            Label: fmt.Sprintf("Order #%s (%s) - %d items - R%.2f", 
                order.OrderNumber, dateStr, numItems, float64(order.Total)/100),
            Value: order.OrderNumber,
        }
    }

    conversation.State.CurrentStep = "track_select_order"
    return "Which order would you like to track?", options, nil
}

func (uc *ProcessMessageUseCase) handleGeneralInquiry(
    ctx context.Context,
    conversation *entities.Conversation,
    extractedData map[string]any,
) (string, []entities.Option, error) {
    // Use AI to generate contextual response
    systemPrompt := `You are a helpful customer support assistant for an e-commerce platform.
    Answer the customer's question politely and concisely. If you don't know the answer, 
    offer to escalate to a human agent. Keep responses under 100 words.`

    response, err := uc.intentClassifier.GenerateResponse(
        ctx,
        conversation.Messages,
        systemPrompt,
    )

    if err != nil {
        return "I apologize, but I'm having trouble processing your request. Would you like me to connect you with a human agent?", 
            []entities.Option{
                {ID: "agent", Label: "Yes, connect me to an agent", Value: "escalate"},
                {ID: "menu", Label: "Show me what you can help with", Value: "show_menu"},
            }, nil
    }

    // Offer follow-up options
    options := []entities.Option{
        {ID: "return", Label: "Return an item", Value: "return"},
        {ID: "track", Label: "Track an order", Value: "track_order"},
        {ID: "agent", Label: "Talk to a person", Value: "escalate"},
    }

    return response + "\n\nIs there anything else I can help you with?", options, nil
}
```

## Integration with Orders API

The chatbot integrates with the Bash Orders API (`https://web-api.bash.com`) to retrieve customer order information. This section documents the client implementation and important considerations when working with the API structure.

### Key API Characteristics

1. **Wrapped Responses**: All responses are wrapped in a standard envelope:
   ```json
   {
     "data": { /* actual data */ },
     "success": true,
     "errorCode": null,
     "errorMessage": null,
     "timestamp": "2026-02-04T11:06:52.846Z",
     "path": "/v1/order/...",
     "request_id": "509dd589-52b1-4247-85b5-477a00223b81"
   }
   ```

2. **Order Number Format**: Orders use format `B{number}-01` (e.g., `B22881153-01`)

3. **Currency**: All monetary values are in **cents** (e.g., `7475` = R74.75)

4. **Nested Items Structure**: Order items are **nested within parcels**, not at the order level:
   ```go
   order.Parcels[0].OrderItems[0].Name
   ```

5. **Multiple Parcels**: Orders can have multiple parcels (deliveries/collections), each with separate tracking

6. **Collection vs Delivery**: 
   - CollectInStore: `shippingAddress` contains store details
   - Courier: `shippingAddress` contains customer delivery address

### Working with Parcel Data

When processing orders for returns or tracking, iterate through parcels to access items:

```go
// Example: Get all items from an order
func GetAllItemsFromOrder(order *DetailedOrder) []OrderItem {
    var allItems []OrderItem
    for _, parcel := range order.Parcels {
        allItems = append(allItems, parcel.OrderItems...)
    }
    return allItems
}

// Example: Get tracking info for all parcels
func GetTrackingInfo(order *DetailedOrder) []TrackingInfo {
    var tracking []TrackingInfo
    for _, parcel := range order.Parcels {
        if parcel.TrackingNumber != "" {
            tracking = append(tracking, TrackingInfo{
                ParcelNumber:   parcel.ParcelNumber,
                TrackingNumber: parcel.TrackingNumber,
                TrackingUrl:    parcel.TrackingUrl,
                Courier:        parcel.Courier,
                Status:         parcel.ParcelStatus,
            })
        }
    }
    return tracking
}

// Example: Build return items list
func BuildReturnItemsList(order *DetailedOrder) []entities.Option {
    var options []entities.Option
    seen := make(map[string]bool)
    
    for _, parcel := range order.Parcels {
        for _, item := range parcel.OrderItems {
            // Deduplicate by SKU (same item may appear multiple times)
            if !seen[item.SKU] {
                seen[item.SKU] = true
                options = append(options, entities.Option{
                    ID:    item.SKU,
                    Label: fmt.Sprintf("%s - R%.2f", item.Name, float64(item.Price)/100),
                    Value: item.SKU,
                })
            }
        }
    }
    return options
}
```

### Orders Client Implementation

```go
package api

import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"
)

type OrdersClient struct {
    baseURL    string
    httpClient *http.Client
}

func NewOrdersClient(baseURL string) *OrdersClient {
    return &OrdersClient{
        baseURL:    baseURL,
        httpClient: &http.Client{Timeout: 10 * time.Second},
    }
}

func (c *OrdersClient) GetOrderHistory(
    ctx context.Context,
    authToken string,
    page, pageSize int,
) (*OrderList, error) {
    url := fmt.Sprintf("%s/v1/order/orders/history?page=%d&pageSize=%d",
        c.baseURL, page, pageSize)

    req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
    if err != nil {
        return nil, err
    }

    req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", authToken))
    req.Header.Set("Content-Type", "application/json")

    resp, err := c.httpClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("orders API returned status %d", resp.StatusCode)
    }

    // API returns wrapped response with metadata
    var apiResponse OrdersAPIResponse
    if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
        return nil, err
    }

    // Check for API-level errors
    if !apiResponse.Success {
        errorMsg := "Unknown error"
        if apiResponse.ErrorMessage != nil {
            errorMsg = *apiResponse.ErrorMessage
        }
        return nil, fmt.Errorf("orders API error: %s", errorMsg)
    }

    return &apiResponse.Data, nil
}

func (c *OrdersClient) GetDetailedOrder(
    ctx context.Context,
    authToken string,
    orderNumber string,
) (*DetailedOrder, error) {
    url := fmt.Sprintf("%s/v1/order/%s", c.baseURL, orderNumber)

    req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
    if err != nil {
        return nil, err
    }

    req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", authToken))
    req.Header.Set("Content-Type", "application/json")

    resp, err := c.httpClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    if resp.StatusCode == http.StatusForbidden {
        return nil, ErrOrderNotOwned
    }

    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("orders API returned status %d", resp.StatusCode)
    }

    // API returns wrapped response with metadata
    var apiResponse DetailedOrderAPIResponse
    if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
        return nil, err
    }

    // Check for API-level errors
    if !apiResponse.Success {
        errorMsg := "Unknown error"
        if apiResponse.ErrorMessage != nil {
            errorMsg = *apiResponse.ErrorMessage
        }
        return nil, fmt.Errorf("orders API error: %s", errorMsg)
    }

    return &apiResponse.Data, nil
}

// Models matching Orders API response structure

// OrdersAPIResponse wraps the actual data with metadata
type OrdersAPIResponse struct {
    Data         OrderList `json:"data"`
    Success      bool      `json:"success"`
    ErrorCode    *string   `json:"errorCode"`
    ErrorMessage *string   `json:"errorMessage"`
    Timestamp    string    `json:"timestamp"`
    Path         string    `json:"path"`
    RequestID    string    `json:"request_id"`
}

type OrderList struct {
    Pages    int            `json:"pages"`
    Total    int            `json:"total"`
    Page     int            `json:"page"`
    PageSize int            `json:"pageSize"`
    Orders   []OrderSummary `json:"orders"`
}

type OrderSummary struct {
    OrderDate              string         `json:"orderDate"`              // e.g., "2025-11-12T20:19:29.867"
    OrderStatus            string         `json:"orderStatus"`            // e.g., "Completed"
    OrderNumber            string         `json:"orderNumber"`            // e.g., "B22881153-01"
    OrderStatusDescription string         `json:"orderStatusDescription"` // e.g., "Complete"
    Total                  int64          `json:"total"`                  // Total in cents
    ItemImageUrls          []string       `json:"itemImageUrls"`          // Product image URLs
    NumberDeliveries       int            `json:"numberDeliveries"`       // Number of separate deliveries
    NumberPendingItems     int            `json:"numberPendingItems"`     // Items not yet fulfilled
    ShippingMethod         ShippingMethod `json:"shippingMethod"`         // Delivery/collection method
}

type ShippingMethod struct {
    ShippingMethodType string `json:"shippingMethodType"` // e.g., "CollectInStore", "Courier"
    ShippingMethodName string `json:"shippingMethodName"` // e.g., "Store Collection"
}

// DetailedOrderAPIResponse wraps the detailed order data with metadata
type DetailedOrderAPIResponse struct {
    Data         DetailedOrder `json:"data"`
    Success      bool          `json:"success"`
    ErrorCode    *string       `json:"errorCode"`
    ErrorMessage *string       `json:"errorMessage"`
    Timestamp    string        `json:"timestamp"`
    Path         string        `json:"path"`
    RequestID    string        `json:"request_id"`
}

type DetailedOrder struct {
    OrderNumber            string         `json:"orderNumber"`            // e.g., "B22881153-01"
    OrderDate              string         `json:"orderDate"`              // e.g., "2025-11-12T20:19:29.867"
    Total                  int64          `json:"total"`                  // Total in cents
    OrderStatus            string         `json:"orderStatus"`            // e.g., "Completed"
    OrderStatusDescription string         `json:"orderStatusDescription"` // e.g., "Complete"
    Parcels                []Parcel       `json:"parcels"`                // Parcel/delivery information
    Invoices               []Invoice      `json:"invoices"`               // Order-level invoices
    Subtotal               int64          `json:"subtotal"`               // Subtotal in cents
    Discount               int64          `json:"discount"`               // Discount in cents
    ShippingCost           int64          `json:"shippingCost"`           // Shipping cost in cents
    DeliveryDescription    *string        `json:"deliveryDescription"`    // Delivery details
    FulfilledIn            string         `json:"fulfilledIn"`            // e.g., "Up to 5 days"
    ItemImageUrls          []string       `json:"itemImageUrls"`          // All item images
    NumberDeliveries       int            `json:"numberDeliveries"`       // Number of parcels
    NumberPendingItems     int            `json:"numberPendingItems"`     // Items not yet delivered
    ShippingMethod         ShippingMethod `json:"shippingMethod"`         // Collection/Delivery type
    ShippingOption         ShippingOption `json:"shippingOption"`         // Standard/Express
    CustomerDetails        CustomerDetails `json:"customerDetails"`       // Customer info
    ShippingAddress        ShippingAddress `json:"shippingAddress"`       // Delivery/collection address
    IsEligibleForReturn    bool           `json:"isEligibleForReturn"`   // Can be returned
}

type Parcel struct {
    ParcelName                string    `json:"parcelName"`                // e.g., "Collection #1"
    ParcelNumber              int       `json:"parcelNumber"`              // Parcel number in sequence
    ParcelStatus              string    `json:"parcelStatus"`              // e.g., "Delivered"
    ParcelStatusHelperText    *string   `json:"parcelStatusHelperText"`   // Additional status info
    ParcelStatusDescription   string    `json:"parcelStatusDescription"`  // Status description
    ParcelInvoices            []Invoice `json:"parcelInvoices"`           // Invoices for this parcel
    ParcelInvoice             *Invoice  `json:"parcelInvoice"`            // Primary invoice
    OrderItems                []OrderItem `json:"orderItems"`             // Items in this parcel
    TrackingNumber            string    `json:"trackingNumber"`           // e.g., "CAZA0747786"
    TrackingUrl               string    `json:"trackingUrl"`              // Tracking URL
    EstimatedArrivalDate      *string   `json:"estimatedArrivalDate"`     // Expected delivery date
    Courier                   string    `json:"courier"`                  // e.g., "bashDelivery"
}

type Invoice struct {
    Url  string `json:"url"`  // Invoice PDF URL
    Date string `json:"date"` // Invoice generation date
}

type OrderItem struct {
    Slug     string  `json:"slug"`     // Product slug/URL identifier
    Name     string  `json:"name"`     // Product name
    Size     string  `json:"size"`     // Size (can be empty)
    Colour   string  `json:"colour"`   // Color/Colour
    Quantity int     `json:"quantity"` // Quantity ordered
    Image    string  `json:"image"`    // Product image URL
    Price    int64   `json:"price"`    // Price in cents
    Store    string  `json:"store"`    // Store name (e.g., "TOTALSPORTS")
    Brand    string  `json:"brand"`    // Brand name
    SKU      string  `json:"sku"`      // Stock keeping unit
    SoldBy   *string `json:"soldBy"`   // Seller (if marketplace)
}

type ShippingOption struct {
    ShippingOptionType string `json:"shippingOptionType"` // e.g., "Standard", "Express"
    ShippingOptionName string `json:"shippingOptionName"` // Display name
}

type CustomerDetails struct {
    Firstname string `json:"firstname"` // First name
    Surname   string `json:"surname"`   // Last name
    Email     string `json:"email"`     // Email address
    Mobile    string `json:"mobile"`    // Mobile number with country code
}

type ShippingAddress struct {
    Line1        string `json:"line1"`        // Address line 1 (e.g., store name)
    Line2        string `json:"line2"`        // Address line 2 (e.g., shop number)
    Line3        string `json:"line3"`        // Address line 3 (e.g., suburb)
    City         string `json:"city"`         // City
    PostalCode   string `json:"postalCode"`   // Postal code
    Name         string `json:"name"`         // Recipient name
    Surname      string `json:"surname"`      // Recipient surname
    MobileNumber string `json:"mobileNumber"` // Recipient mobile
    Email        string `json:"email"`        // Recipient email
}

var ErrOrderNotOwned = fmt.Errorf("customer does not own this order")
```

## Authentication & Authorization

### JWT Middleware

```go
package middleware

import (
    "context"
    "net/http"
    "strings"

    "github.com/gin-gonic/gin"
    "github.com/golang-jwt/jwt/v5"
)

type CustomerAuthMiddleware struct {
    jwtSecret []byte
}

func NewCustomerAuthMiddleware(secret string) *CustomerAuthMiddleware {
    return &CustomerAuthMiddleware{
        jwtSecret: []byte(secret),
    }
}

func (m *CustomerAuthMiddleware) Authenticate() gin.HandlerFunc {
    return func(c *gin.Context) {
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
            c.Abort()
            return
        }

        tokenString := strings.TrimPrefix(authHeader, "Bearer ")
        if tokenString == authHeader {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization format"})
            c.Abort()
            return
        }

        token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
            if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
                return nil, fmt.Errorf("unexpected signing method")
            }
            return m.jwtSecret, nil
        })

        if err != nil || !token.Valid {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
            c.Abort()
            return
        }

        claims, ok := token.Claims.(jwt.MapClaims)
        if !ok {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token claims"})
            c.Abort()
            return
        }

        // Extract customer info from token
        customerID, ok := claims["sub"].(string)
        if !ok {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid customer ID in token"})
            c.Abort()
            return
        }

        email, _ := claims["email"].(string)

        // Set in context for downstream handlers
        c.Set("customer_id", customerID)
        c.Set("customer_email", email)
        c.Set("auth_token", tokenString)

        c.Next()
    }
}

// Helper to extract customer ID from context
func GetCustomerID(c *gin.Context) string {
    customerID, _ := c.Get("customer_id")
    return customerID.(string)
}

func GetAuthToken(c *gin.Context) string {
    token, _ := c.Get("auth_token")
    return token.(string)
}
```

## AI Integration for Natural Language Processing

### Intent Classifier

```go
package ai

import (
    "context"
    "encoding/json"

    "github.com/sashabaranov/go-openai"
)

type IntentClassifier struct {
    client *openai.Client
}

type ClassificationResult struct {
    Intent     string         `json:"intent"`
    Confidence float64        `json:"confidence"`
    Entities   map[string]any `json:"entities"`
}

func NewIntentClassifier(apiKey string) *IntentClassifier {
    return &IntentClassifier{
        client: openai.NewClient(apiKey),
    }
}

func (ic *IntentClassifier) Classify(
    ctx context.Context,
    message string,
    conversationHistory []entities.Message,
) (string, map[string]any, float64) {
    // Build conversation context
    messages := []openai.ChatCompletionMessage{
        {
            Role: openai.ChatMessageRoleSystem,
            Content: `You are an intent classifier for a customer support chatbot.
Analyze the user's message and identify their intent based on what they're trying to accomplish.

Supported Intents (Phase 1 - Active):
- return: Customer wants to return purchased items
  * Triggers: "return", "send back", "don't want", "refund", "wrong size", "defective"
- track_order: Customer wants to track order delivery
  * Triggers: "where is my order", "track", "delivery status", "when will it arrive"
- general_inquiry: General questions, fallback intent
  * Triggers: Any other customer service question

Future Intents (Phase 2 - Recognition only, respond with "This feature is coming soon"):
- refund_status: Check refund processing status
  * Triggers: "refund status", "where's my refund", "money back"
- change_address: Update delivery address
  * Triggers: "change address", "wrong address", "update location"
- cancel_order: Cancel pending order
  * Triggers: "cancel order", "cancel my purchase", "don't want anymore"
- password_reset: Reset account password
  * Triggers: "forgot password", "reset password", "can't login"
- update_profile: Update account information
  * Triggers: "change email", "update phone", "edit profile"

Entity Extraction:
- order_number: Any order number mentioned (format: #12345 or 12345)
- product_mention: Specific product/item names
- reason_mention: Reason for return/issue/complaint
- address_mention: Any address information
- urgency_level: low, medium, high (based on language sentiment)

Respond with JSON only:
{
  "intent": "return",
  "confidence": 0.95,
  "entities": {
    "product_mention": "Nike shoes",
    "reason_mention": "too small",
    "urgency_level": "medium"
  },
  "clarification_needed": false,
  "suggested_response": "I can help you return your Nike shoes."
}`,
        },
    }

    // Add conversation history (last 5 messages for context)
    historyStart := len(conversationHistory) - 5
    if historyStart < 0 {
        historyStart = 0
    }
    for _, msg := range conversationHistory[historyStart:] {
        role := openai.ChatMessageRoleUser
        if msg.Role == entities.MessageRoleAssistant {
            role = openai.ChatMessageRoleAssistant
        }
        messages = append(messages, openai.ChatCompletionMessage{
            Role:    role,
            Content: msg.Content,
        })
    }

    // Add current message
    messages = append(messages, openai.ChatCompletionMessage{
        Role:    openai.ChatMessageRoleUser,
        Content: message,
    })

    resp, err := ic.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
        Model:       openai.GPT4,
        Messages:    messages,
        Temperature: 0.3,
    })

    if err != nil {
        // Fallback to general_inquiry
        return "general_inquiry", make(map[string]any), 0.5
    }

    var result ClassificationResult
    if err := json.Unmarshal([]byte(resp.Choices[0].Message.Content), &result); err != nil {
        return "general_inquiry", make(map[string]any), 0.5
    }

    return result.Intent, result.Entities, result.Confidence
}

func (ic *IntentClassifier) GenerateResponse(
    ctx context.Context,
    conversationHistory []entities.Message,
    systemPrompt string,
) (string, error) {
    messages := []openai.ChatCompletionMessage{
        {
            Role:    openai.ChatMessageRoleSystem,
            Content: systemPrompt,
        },
    }

    for _, msg := range conversationHistory {
        role := openai.ChatMessageRoleUser
        if msg.Role == entities.MessageRoleAssistant {
            role = openai.ChatMessageRoleAssistant
        }
        messages = append(messages, openai.ChatCompletionMessage{
            Role:    role,
            Content: msg.Content,
        })
    }

    resp, err := ic.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
        Model:       openai.GPT4,
        Messages:    messages,
        Temperature: 0.7,
    })

    if err != nil {
        return "", err
    }

    return resp.Choices[0].Message.Content, nil
}
```

## Database Schema

### SQLite Version (Local Development - No Docker Required)

**migrations/001_sqlite_schema.sql**:

```sql
-- Conversations table
CREATE TABLE conversations (
    id TEXT PRIMARY KEY, -- UUID as string
    customer_id TEXT NOT NULL,
    type TEXT NOT NULL CHECK (type IN ('natural_language', 'guided_flow')),
    intent TEXT,
    current_step TEXT,
    flow_id TEXT,
    context TEXT NOT NULL DEFAULT '{}', -- JSON stored as TEXT, marshal/unmarshal in Go
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    completed_at TEXT
);

CREATE INDEX idx_conversations_customer_id ON conversations(customer_id);
CREATE INDEX idx_conversations_created_at ON conversations(created_at);

-- Messages table
CREATE TABLE messages (
    id TEXT PRIMARY KEY,
    conversation_id TEXT NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    role TEXT NOT NULL CHECK (role IN ('user', 'assistant', 'system')),
    content TEXT NOT NULL,
    metadata TEXT NOT NULL DEFAULT '{}', -- JSON as TEXT
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_messages_conversation_id ON messages(conversation_id);
CREATE INDEX idx_messages_created_at ON messages(created_at);

-- Return requests table
CREATE TABLE return_requests (
    id TEXT PRIMARY KEY,
    conversation_id TEXT REFERENCES conversations(id) ON DELETE SET NULL,
    customer_id TEXT NOT NULL,
    order_number TEXT NOT NULL,
    items TEXT NOT NULL, -- JSON array as TEXT
    reason TEXT NOT NULL,
    detailed_reason TEXT,
    photos TEXT, -- JSON array as TEXT
    refund_method TEXT NOT NULL,
    delivery_method TEXT NOT NULL,
    collection_point TEXT, -- JSON object as TEXT
    shipping_address TEXT, -- JSON object as TEXT
    status TEXT NOT NULL DEFAULT 'draft',
    estimated_refund_date TEXT,
    refund_amount_cents INTEGER NOT NULL,
    return_reference TEXT,
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_return_requests_customer_id ON return_requests(customer_id);
CREATE INDEX idx_return_requests_order_number ON return_requests(order_number);
CREATE INDEX idx_return_requests_status ON return_requests(status);
CREATE INDEX idx_return_requests_created_at ON return_requests(created_at);

-- Refund requests table (Phase 2)
CREATE TABLE refund_requests (
    id TEXT PRIMARY KEY,
    conversation_id TEXT REFERENCES conversations(id) ON DELETE SET NULL,
    customer_id TEXT NOT NULL,
    order_number TEXT NOT NULL,
    return_reference TEXT,
    refund_amount_cents INTEGER NOT NULL,
    refund_method TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending',
    processed_date TEXT,
    failure_reason TEXT,
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_refund_requests_customer_id ON refund_requests(customer_id);
CREATE INDEX idx_refund_requests_order_number ON refund_requests(order_number);
CREATE INDEX idx_refund_requests_status ON refund_requests(status);
CREATE INDEX idx_refund_requests_return_reference ON refund_requests(return_reference);

-- Account actions table (Phase 2)
CREATE TABLE account_actions (
    id TEXT PRIMARY KEY,
    conversation_id TEXT REFERENCES conversations(id) ON DELETE SET NULL,
    customer_id TEXT NOT NULL,
    action_type TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending',
    otp_code TEXT,
    otp_expires_at TEXT,
    otp_attempts INTEGER DEFAULT 0,
    action_data TEXT NOT NULL DEFAULT '{}', -- JSON as TEXT
    verified_at TEXT,
    completed_at TEXT,
    failure_reason TEXT,
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_account_actions_customer_id ON account_actions(customer_id);
CREATE INDEX idx_account_actions_status ON account_actions(status);
CREATE INDEX idx_account_actions_action_type ON account_actions(action_type);
CREATE INDEX idx_account_actions_created_at ON account_actions(created_at);

-- Flows table (for API-configurable flows)
CREATE TABLE flows (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    intent TEXT NOT NULL,
    version INTEGER NOT NULL DEFAULT 1,
    steps TEXT NOT NULL, -- JSON array as TEXT
    is_active INTEGER NOT NULL DEFAULT 0, -- BOOLEAN as INTEGER (0/1)
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_flows_intent ON flows(intent);
CREATE INDEX idx_flows_is_active ON flows(is_active);
CREATE UNIQUE INDEX idx_flows_active_intent ON flows(intent) WHERE is_active = 1;

-- Audit log for tracking changes
CREATE TABLE audit_logs (
    id TEXT PRIMARY KEY,
    entity_type TEXT NOT NULL,
    entity_id TEXT NOT NULL,
    action TEXT NOT NULL,
    user_id TEXT,
    changes TEXT, -- JSON as TEXT
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_audit_logs_entity ON audit_logs(entity_type, entity_id);
CREATE INDEX idx_audit_logs_created_at ON audit_logs(created_at);
```

**Notes on SQLite Schema:**
- UUIDs stored as TEXT (generate with `uuid.New().String()` in Go)
- JSONB columns stored as TEXT (use `json.Marshal()`/`json.Unmarshal()` in Go)
- TIMESTAMP stored as TEXT in ISO8601 format (`time.Now().Format(time.RFC3339)`)
- BOOLEAN stored as INTEGER (0 = false, 1 = true)
- BIGINT stored as INTEGER (SQLite's INTEGER is 64-bit)

### PostgreSQL Version (Production)

For production with Docker, see **migrations/001_postgres_schema.sql** with:
- Native UUID types with `uuid_generate_v4()`
- JSONB columns for efficient JSON querying
- GIN indexes on JSONB fields
- TIMESTAMP WITH TIME ZONE
- TEXT[] arrays

## Configuration

### SQLite Configuration (Local Development)

```yaml
# config/config.yaml

server:
  port: 8080
  environment: development
  read_timeout: 30s
  write_timeout: 30s
  shutdown_timeout: 10s

database:
  type: sqlite
  path: ./data/chatbot.db
  max_open_conns: 1  # SQLite limitation
  max_idle_conns: 1

cache:
  type: memory  # In-memory cache instead of Redis
  session_ttl: 1h
  max_entries: 10000

auth:
  jwt_secret: ${JWT_SECRET}
  token_expiration: 24h

integrations:
  orders_api:
    base_url: https://web-api.bash.com
    timeout: 10s
  return_management_api:
    base_url: ${RETURN_MGMT_API_URL}
    api_key: ${RETURN_MGMT_API_KEY}
    timeout: 15s

ai:
  provider: keyword  # Use keyword-based for MVP (no API key needed)
  # provider: openai  # Switch to this when ready
  # api_key: ${AI_API_KEY}

storage:
  provider: local  # Local file storage for photos
  path: ./data/uploads

logging:
  level: debug
  format: json
```

### PostgreSQL Configuration (Production)

```yaml
# config/config.yaml

server:
  port: 8080
  environment: production
  read_timeout: 30s
  write_timeout: 30s
  shutdown_timeout: 10s

database:
  type: postgres
  host: ${DB_HOST}
  port: 5432
  name: ${DB_NAME}
  user: ${DB_USER}
  password: ${DB_PASSWORD}
  max_open_conns: 25
  max_idle_conns: 5
  conn_max_lifetime: 5m

redis:
  host: ${REDIS_HOST}
  port: 6379
  password: ${REDIS_PASSWORD}
  db: 0
  session_ttl: 1h

auth:
  jwt_secret: ${JWT_SECRET}
  token_expiration: 24h

integrations:
  orders_api:
    base_url: https://web-api.bash.com  # Bash Orders API base URL
    timeout: 10s
  return_management_api:
    base_url: ${RETURN_MGMT_API_URL}
    api_key: ${RETURN_MGMT_API_KEY}
    timeout: 15s

ai:
  provider: openai # or anthropic
  api_key: ${AI_API_KEY}
  model: gpt-4
  max_tokens: 1000
  temperature: 0.7

storage:
  provider: s3 # for photo uploads
  bucket: ${S3_BUCKET}
  region: ${AWS_REGION}
  access_key: ${AWS_ACCESS_KEY}
  secret_key: ${AWS_SECRET_KEY}

logging:
  level: info
  format: json

monitoring:
  enabled: true
  prometheus_port: 9090
```

## Deployment

### Dockerfile

```dockerfile
# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/api

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/main .
COPY --from=builder /app/config ./config

EXPOSE 8080

CMD ["./main"]
```

### Kubernetes Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: customer-support-api
spec:
  replicas: 3
  selector:
    matchLabels:
      app: customer-support-api
  template:
    metadata:
      labels:
        app: customer-support-api
    spec:
      containers:
      - name: api
        image: customer-support-api:latest
        ports:
        - containerPort: 8080
        env:
        - name: DB_HOST
          valueFrom:
            secretKeyRef:
              name: db-secrets
              key: host
        - name: DB_PASSWORD
          valueFrom:
            secretKeyRef:
              name: db-secrets
              key: password
        - name: JWT_SECRET
          valueFrom:
            secretKeyRef:
              name: auth-secrets
              key: jwt-secret
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /ready
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
```

## Testing Strategy

### MVP/POC Testing Approach

For the MVP, focus on these testing layers:

1. **Unit Tests** - Core business logic (use cases, services)
2. **Integration Tests** - API endpoints with test database
3. **E2E Tests** - Full user flows with mocked external services
4. **Contract Tests** - Orders API integration with recorded responses

### Test Data Setup

#### Mock Orders API Responses

Store these as fixtures in `tests/fixtures/orders_api/`:

**`order_history_response.json`**:
```json
{
  "data": {
    "pages": 1,
    "total": 2,
    "page": 0,
    "pageSize": 20,
    "orders": [
      {
        "orderDate": "2025-11-12T20:19:29.867",
        "orderStatus": "Completed",
        "orderNumber": "B22881153-01",
        "orderStatusDescription": "Complete",
        "total": 7475,
        "itemImageUrls": ["https://assets.bash.com/products/1000x1000/57206641.jpg"],
        "numberDeliveries": 1,
        "numberPendingItems": 0,
        "shippingMethod": {
          "shippingMethodType": "CollectInStore",
          "shippingMethodName": "Store Collection"
        }
      }
    ]
  },
  "success": true,
  "errorCode": null,
  "errorMessage": null,
  "timestamp": "2026-02-04T11:04:07.036Z",
  "path": "/v1/order/orders/history",
  "request_id": "test-request-id"
}
```

**`order_detail_response.json`**:
```json
{
  "data": {
    "orderNumber": "B22881153-01",
    "orderDate": "2025-11-12T20:19:29.867",
    "total": 7475,
    "orderStatus": "Completed",
    "parcels": [
      {
        "parcelName": "Collection #1",
        "parcelNumber": 1,
        "parcelStatus": "Delivered",
        "orderItems": [
          {
            "slug": "test-product",
            "name": "Test Product",
            "size": "M",
            "colour": "Blue",
            "quantity": 1,
            "image": "https://assets.bash.com/products/test.jpg",
            "price": 7475,
            "store": "BASH",
            "brand": "TestBrand",
            "sku": "TEST123",
            "soldBy": null
          }
        ],
        "trackingNumber": "TRACK123",
        "trackingUrl": "https://track.example.com/TRACK123",
        "courier": "bashDelivery"
      }
    ],
    "subtotal": 7475,
    "discount": 0,
    "shippingCost": 0,
    "shippingMethod": {
      "shippingMethodType": "CollectInStore",
      "shippingMethodName": "Store Collection"
    },
    "customerDetails": {
      "firstname": "Test",
      "surname": "User",
      "email": "test@example.com",
      "mobile": "+27721234567"
    },
    "shippingAddress": {
      "line1": "Test Store",
      "city": "Cape Town",
      "postalCode": "8001"
    },
    "isEligibleForReturn": true
  },
  "success": true
}
```

#### Database Seed Data

**Initial Flow Configuration** (`migrations/seeds/001_initial_flows.sql`):

```sql
-- Insert Returns Flow
INSERT INTO flows (id, name, intent, version, steps, is_active, created_at, updated_at)
VALUES (
  '550e8400-e29b-41d4-a716-446655440000',
  'Standard Returns Flow',
  'return',
  1,
  '[
    {
      "id": "identify_order",
      "type": "single_choice",
      "prompt": "Which order contains the item you would like to return?",
      "options": [],
      "next_steps": {"default": "select_items"}
    },
    {
      "id": "select_items",
      "type": "multi_choice",
      "prompt": "Which item(s) would you like to return?",
      "options": [],
      "next_steps": {"default": "select_reason"}
    },
    {
      "id": "select_reason",
      "type": "single_choice",
      "prompt": "What is the reason for the return?",
      "options": [
        {"id": "wrong_size", "label": "Wrong size", "value": "wrong_size"},
        {"id": "quality", "label": "Quality issue", "value": "quality_issue"},
        {"id": "changed_mind", "label": "Changed my mind", "value": "changed_mind"},
        {"id": "other", "label": "Other", "value": "other"}
      ],
      "next_steps": {"other": "detailed_reason", "default": "select_refund"}
    },
    {
      "id": "detailed_reason",
      "type": "text_input",
      "prompt": "Please provide more details",
      "validation": {"type": "min_length", "value": "10"},
      "next_steps": {"default": "select_refund"}
    },
    {
      "id": "select_refund",
      "type": "single_choice",
      "prompt": "How would you like to receive your refund?",
      "options": [
        {"id": "original", "label": "Original payment method", "value": "original_payment"},
        {"id": "credit", "label": "Store credit", "value": "store_credit"}
      ],
      "next_steps": {"default": "confirmation"}
    },
    {
      "id": "confirmation",
      "type": "confirmation",
      "prompt": "Please review your return request",
      "is_terminal": true,
      "next_steps": {}
    }
  ]'::jsonb,
  true,
  NOW(),
  NOW()
);
```

### End-to-End Test Scenarios

#### E2E Test 1: Natural Language Return Flow

```go
package e2e_test

import (
    "context"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "strings"
    "testing"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/suite"
)

type ReturnFlowE2ETestSuite struct {
    suite.Suite
    router         *gin.Engine
    db             *sql.DB
    ordersAPIMock  *httptest.Server
    authToken      string
}

func (suite *ReturnFlowE2ETestSuite) SetupSuite() {
    // Setup test database
    suite.db = setupTestDB(suite.T())
    
    // Run migrations
    runMigrations(suite.db)
    
    // Seed initial flow data
    seedFlows(suite.db)
    
    // Setup mock Orders API
    suite.ordersAPIMock = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        suite.handleMockOrdersAPI(w, r)
    }))
    
    // Configure app with mock API URL
    config := &Config{
        OrdersAPIBaseURL: suite.ordersAPIMock.URL,
    }
    
    // Initialize router
    suite.router = setupRouter(suite.db, config)
    
    // Generate test auth token
    suite.authToken = generateTestJWT("test-customer-id", "test@example.com")
}

func (suite *ReturnFlowE2ETestSuite) TearDownSuite() {
    suite.ordersAPIMock.Close()
    suite.db.Close()
}

func (suite *ReturnFlowE2ETestSuite) handleMockOrdersAPI(w http.ResponseWriter, r *http.Request) {
    switch {
    case strings.HasSuffix(r.URL.Path, "/orders/history"):
        // Return order history fixture
        fixture := loadFixture("order_history_response.json")
        w.Header().Set("Content-Type", "application/json")
        w.Write(fixture)
        
    case strings.Contains(r.URL.Path, "/order/B22881153-01"):
        // Return order detail fixture
        fixture := loadFixture("order_detail_response.json")
        w.Header().Set("Content-Type", "application/json")
        w.Write(fixture)
        
    default:
        w.WriteHeader(http.StatusNotFound)
    }
}

func (suite *ReturnFlowE2ETestSuite) TestCompleteReturnFlow_NaturalLanguage() {
    // Step 1: Start conversation
    conversationResp := suite.makeRequest("POST", "/api/v1/conversations", map[string]any{
        "type":            "natural_language",
        "initial_message": "I want to return my Nike shoes. They're too small",
    })
    
    assert.Equal(suite.T(), http.StatusCreated, conversationResp.StatusCode)
    
    var convData map[string]any
    json.Unmarshal(conversationResp.Body, &convData)
    conversationID := convData["conversation_id"].(string)
    
    assert.Equal(suite.T(), "return", convData["intent"])
    assert.NotEmpty(suite.T(), conversationID)
    
    // Step 2: Provide order number
    msgResp := suite.makeRequest("POST", fmt.Sprintf("/api/v1/conversations/%s/messages", conversationID), map[string]any{
        "content": "Order B22881153-01",
    })
    
    assert.Equal(suite.T(), http.StatusOK, msgResp.StatusCode)
    
    var msgData map[string]any
    json.Unmarshal(msgResp.Body, &msgData)
    
    // Should show items from order
    message := msgData["message"].(map[string]any)
    assert.Contains(suite.T(), message["content"], "Test Product")
    
    // Step 3: Select item to return
    msgResp = suite.makeRequest("POST", fmt.Sprintf("/api/v1/conversations/%s/messages", conversationID), map[string]any{
        "content": "The Test Product",
    })
    
    assert.Equal(suite.T(), http.StatusOK, msgResp.StatusCode)
    
    // Step 4: Provide return reason
    msgResp = suite.makeRequest("POST", fmt.Sprintf("/api/v1/conversations/%s/messages", conversationID), map[string]any{
        "content": "Wrong size",
    })
    
    assert.Equal(suite.T(), http.StatusOK, msgResp.StatusCode)
    
    // Step 5: Select refund method
    msgResp = suite.makeRequest("POST", fmt.Sprintf("/api/v1/conversations/%s/messages", conversationID), map[string]any{
        "content": "Original payment method",
    })
    
    assert.Equal(suite.T(), http.StatusOK, msgResp.StatusCode)
    
    // Step 6: Create return request
    returnResp := suite.makeRequest("POST", "/api/v1/returns", map[string]any{
        "conversation_id": conversationID,
        "order_number":    "B22881153-01",
        "items": []map[string]any{
            {"item_id": "TEST123", "quantity": 1},
        },
        "reason":         "wrong_size",
        "refund_method":  "original_payment",
        "delivery_method": "courier",
        "shipping_address": map[string]any{
            "street":      "123 Test St",
            "city":        "Cape Town",
            "province":    "Western Cape",
            "postal_code": "8001",
            "country":     "ZA",
        },
    })
    
    assert.Equal(suite.T(), http.StatusCreated, returnResp.StatusCode)
    
    var returnData map[string]any
    json.Unmarshal(returnResp.Body, &returnData)
    
    assert.NotEmpty(suite.T(), returnData["id"])
    assert.NotEmpty(suite.T(), returnData["return_reference"])
    assert.Equal(suite.T(), "pending_approval", returnData["status"])
    assert.Equal(suite.T(), int64(7475), returnData["refund_amount_cents"])
}

func (suite *ReturnFlowE2ETestSuite) TestCompleteReturnFlow_GuidedFlow() {
    // Step 1: Initiate guided return flow
    flowResp := suite.makeRequest("POST", "/api/v1/conversations", map[string]any{
        "type":   "guided_flow",
        "intent": "return",
    })
    
    assert.Equal(suite.T(), http.StatusCreated, flowResp.StatusCode)
    
    var flowData map[string]any
    json.Unmarshal(flowResp.Body, &flowData)
    conversationID := flowData["conversation_id"].(string)
    
    // Step 2: Select order (flow action)
    actionResp := suite.makeRequest("POST", fmt.Sprintf("/api/v1/conversations/%s/flow-action", conversationID), map[string]any{
        "action": "select_option",
        "value":  "B22881153-01",
    })
    
    assert.Equal(suite.T(), http.StatusOK, actionResp.StatusCode)
    
    // Continue through flow steps...
    // (Similar to natural language but using flow-action endpoint)
}

func (suite *ReturnFlowE2ETestSuite) makeRequest(method, path string, body any) *Response {
    jsonBody, _ := json.Marshal(body)
    req := httptest.NewRequest(method, path, strings.NewReader(string(jsonBody)))
    req.Header.Set("Authorization", "Bearer "+suite.authToken)
    req.Header.Set("Content-Type", "application/json")
    
    w := httptest.NewRecorder()
    suite.router.ServeHTTP(w, req)
    
    return &Response{
        StatusCode: w.Code,
        Body:       w.Body.Bytes(),
        Headers:    w.Header(),
    }
}

func TestReturnFlowE2ESuite(t *testing.T) {
    suite.Run(t, new(ReturnFlowE2ETestSuite))
}
```

#### E2E Test 2: Order History & Tracking

```go
func (suite *ReturnFlowE2ETestSuite) TestOrderHistoryAndTracking() {
    // Test order history endpoint
    historyResp := suite.makeRequest("GET", "/api/v1/orders?page=0&page_size=20", nil)
    
    assert.Equal(suite.T(), http.StatusOK, historyResp.StatusCode)
    
    var historyData map[string]any
    json.Unmarshal(historyResp.Body, &historyData)
    
    orders := historyData["orders"].([]any)
    assert.Greater(suite.T(), len(orders), 0)
    
    // Test order detail endpoint
    orderResp := suite.makeRequest("GET", "/api/v1/orders/B22881153-01", nil)
    
    assert.Equal(suite.T(), http.StatusOK, orderResp.StatusCode)
    
    var orderData map[string]any
    json.Unmarshal(orderResp.Body, &orderData)
    
    assert.Equal(suite.T(), "B22881153-01", orderData["orderNumber"])
    assert.NotEmpty(suite.T(), orderData["parcels"])
}
```

### Test Helper Functions

```go
// tests/helpers/fixtures.go
package helpers

import (
    "os"
    "path/filepath"
)

func LoadFixture(filename string) []byte {
    path := filepath.Join("tests", "fixtures", "orders_api", filename)
    data, err := os.ReadFile(path)
    if err != nil {
        panic(err)
    }
    return data
}

// tests/helpers/auth.go
package helpers

import (
    "time"
    
    "github.com/golang-jwt/jwt/v5"
)

func GenerateTestJWT(customerID, email string) string {
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
        "sub":   customerID,
        "email": email,
        "exp":   time.Now().Add(24 * time.Hour).Unix(),
    })
    
    tokenString, _ := token.SignedString([]byte("test-secret"))
    return tokenString
}

// tests/helpers/database.go
package helpers

import (
    "database/sql"
    "testing"
    
    _ "github.com/lib/pq"
)

func SetupTestDB(t *testing.T) *sql.DB {
    db, err := sql.Open("postgres", "postgres://localhost/chatbot_test?sslmode=disable")
    if err != nil {
        t.Fatalf("Failed to connect to test database: %v", err)
    }
    
    // Clean database before each test
    cleanDatabase(db)
    
    return db
}

func CleanDatabase(db *sql.DB) {
    db.Exec("TRUNCATE conversations, messages, return_requests, flows CASCADE")
}
```

### Unit Tests

```go
package usecases_test

import (
    "context"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
)

func TestProcessMessage_ReturnIntent(t *testing.T) {
    // Arrange
    mockConversationService := new(MockConversationService)
    mockIntentClassifier := new(MockIntentClassifier)
    mockOrderService := new(MockOrderService)

    useCase := &ProcessMessageUseCase{
        conversationService: mockConversationService,
        intentClassifier:    mockIntentClassifier,
        orderService:        mockOrderService,
    }

    conversationID := "test-conversation-id"
    userMessage := "I want to return my shoes"

    conversation := &entities.Conversation{
        ID:         conversationID,
        CustomerID: "customer-123",
        Messages:   []entities.Message{},
    }

    mockConversationService.On("GetByID", mock.Anything, conversationID).
        Return(conversation, nil)

    mockIntentClassifier.On("Classify", mock.Anything, userMessage, mock.Anything).
        Return("return", map[string]any{"product_mention": "shoes"}, 0.95)

    // Act
    response, err := useCase.Execute(context.Background(), conversationID, userMessage)

    // Assert
    assert.NoError(t, err)
    assert.NotNil(t, response)
    assert.Equal(t, "return", conversation.Intent)
    mockConversationService.AssertExpectations(t)
    mockIntentClassifier.AssertExpectations(t)
}
```

### Integration Tests

```go
package integration_test

import (
    "context"
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/gin-gonic/gin"
    "github.com/stretchr/testify/assert"
)

func TestConversationEndpoint_CreateConversation(t *testing.T) {
    // Setup test database and router
    db := setupTestDB(t)
    defer db.Close()

    router := setupTestRouter(db)

    // Create request
    w := httptest.NewRecorder()
    req, _ := http.NewRequest("POST", "/api/v1/conversations", strings.NewReader(`{
        "type": "natural_language",
        "initial_message": "I want to return my shoes"
    }`))
    req.Header.Set("Authorization", "Bearer "+generateTestToken())
    req.Header.Set("Content-Type", "application/json")

    // Execute request
    router.ServeHTTP(w, req)

    // Assert response
    assert.Equal(t, http.StatusCreated, w.Code)

    var response map[string]interface{}
    json.Unmarshal(w.Body.Bytes(), &response)

    assert.NotEmpty(t, response["conversation_id"])
    assert.Equal(t, "return", response["intent"])
}
```

## Comprehensive Test Cases

This section provides detailed test cases covering all critical functionality for the MVP.

### 1. Unit Tests - Core Business Logic

#### 1.1 Order Item Extraction from Parcels

**Critical**: Items are nested in parcels, not at order level. These tests ensure proper extraction.

```go
package domain_test

import (
    "testing"
    
    "github.com/stretchr/testify/assert"
)

func TestGetAllItemsFromOrder_SingleParcel(t *testing.T) {
    // Given: Order with one parcel containing 2 items
    order := &DetailedOrder{
        OrderNumber: "B22881153-01",
        Parcels: []Parcel{
            {
                ParcelName: "Collection #1",
                OrderItems: []OrderItem{
                    {SKU: "TEST001", Name: "Nike Shoes", Quantity: 1, Price: 5000},
                    {SKU: "TEST002", Name: "Adidas Shirt", Quantity: 2, Price: 1500},
                },
            },
        },
    }
    
    // When: Extracting all items
    items := GetAllItemsFromOrder(order)
    
    // Then: Should return all items from the parcel
    assert.Len(t, items, 2)
    assert.Equal(t, "TEST001", items[0].SKU)
    assert.Equal(t, "TEST002", items[1].SKU)
}

func TestGetAllItemsFromOrder_MultipleParcels(t *testing.T) {
    // Given: Order with multiple parcels
    order := &DetailedOrder{
        OrderNumber: "B22881153-01",
        Parcels: []Parcel{
            {
                ParcelName: "Delivery #1",
                OrderItems: []OrderItem{
                    {SKU: "TEST001", Name: "Item 1", Quantity: 1, Price: 5000},
                },
            },
            {
                ParcelName: "Delivery #2",
                OrderItems: []OrderItem{
                    {SKU: "TEST002", Name: "Item 2", Quantity: 1, Price: 3000},
                    {SKU: "TEST003", Name: "Item 3", Quantity: 1, Price: 2000},
                },
            },
        },
    }
    
    // When: Extracting all items
    items := GetAllItemsFromOrder(order)
    
    // Then: Should return items from all parcels
    assert.Len(t, items, 3)
    assert.Equal(t, "TEST001", items[0].SKU)
    assert.Equal(t, "TEST002", items[1].SKU)
    assert.Equal(t, "TEST003", items[2].SKU)
}

func TestGetAllItemsFromOrder_EmptyParcels(t *testing.T) {
    // Given: Order with no parcels
    order := &DetailedOrder{
        OrderNumber: "B22881153-01",
        Parcels:     []Parcel{},
    }
    
    // When: Extracting items
    items := GetAllItemsFromOrder(order)
    
    // Then: Should return empty slice
    assert.Empty(t, items)
}
```

#### 1.2 Return Eligibility Validation

```go
func TestIsOrderEligibleForReturn_HappyPath(t *testing.T) {
    // Given: Completed order within 30 days
    order := &DetailedOrder{
        OrderNumber:         "B22881153-01",
        OrderStatus:         "Completed",
        IsEligibleForReturn: true,
        OrderDate:           time.Now().Add(-10 * 24 * time.Hour).Format(time.RFC3339),
    }
    
    // When: Checking eligibility
    eligible, reason := IsOrderEligibleForReturn(order)
    
    // Then: Should be eligible
    assert.True(t, eligible)
    assert.Empty(t, reason)
}

func TestIsOrderEligibleForReturn_NotCompleted(t *testing.T) {
    // Given: Order in "Processing" status
    order := &DetailedOrder{
        OrderNumber:         "B22881153-01",
        OrderStatus:         "Processing",
        IsEligibleForReturn: false,
        OrderDate:           time.Now().Add(-5 * 24 * time.Hour).Format(time.RFC3339),
    }
    
    // When: Checking eligibility
    eligible, reason := IsOrderEligibleForReturn(order)
    
    // Then: Should NOT be eligible
    assert.False(t, eligible)
    assert.Contains(t, reason, "not yet delivered")
}

func TestIsOrderEligibleForReturn_TooOld(t *testing.T) {
    // Given: Order completed 45 days ago (beyond 30-day window)
    order := &DetailedOrder{
        OrderNumber:         "B22881153-01",
        OrderStatus:         "Completed",
        IsEligibleForReturn: false,
        OrderDate:           time.Now().Add(-45 * 24 * time.Hour).Format(time.RFC3339),
    }
    
    // When: Checking eligibility
    eligible, reason := IsOrderEligibleForReturn(order)
    
    // Then: Should NOT be eligible
    assert.False(t, eligible)
    assert.Contains(t, reason, "30-day return window")
}

func TestIsOrderEligibleForReturn_AlreadyReturned(t *testing.T) {
    // Given: Order with existing return
    order := &DetailedOrder{
        OrderNumber:         "B22881153-01",
        OrderStatus:         "Returned",
        IsEligibleForReturn: false,
    }
    
    // When: Checking eligibility
    eligible, reason := IsOrderEligibleForReturn(order)
    
    // Then: Should NOT be eligible
    assert.False(t, eligible)
    assert.Contains(t, reason, "already returned")
}
```

#### 1.3 Intent Classification (Keyword-based for MVP)

```go
func TestKeywordClassifier_ReturnIntent(t *testing.T) {
    classifier := NewKeywordClassifier()
    
    testCases := []struct {
        message  string
        expected string
    }{
        {"I want to return my shoes", "return"},
        {"Can I send back this item?", "return"},
        {"This doesn't fit, I need a refund", "return"},
        {"Wrong size, need to return", "return"},
        {"Item is defective, return please", "return"},
    }
    
    for _, tc := range testCases {
        t.Run(tc.message, func(t *testing.T) {
            intent, _, confidence := classifier.Classify(context.Background(), tc.message, nil)
            assert.Equal(t, tc.expected, intent)
            assert.Greater(t, confidence, 0.7)
        })
    }
}

func TestKeywordClassifier_TrackOrderIntent(t *testing.T) {
    classifier := NewKeywordClassifier()
    
    testCases := []struct {
        message  string
        expected string
    }{
        {"Where is my order?", "track_order"},
        {"Track my delivery", "track_order"},
        {"When will my package arrive?", "track_order"},
        {"Order status please", "track_order"},
        {"Has my order shipped?", "track_order"},
    }
    
    for _, tc := range testCases {
        t.Run(tc.message, func(t *testing.T) {
            intent, _, _ := classifier.Classify(context.Background(), tc.message, nil)
            assert.Equal(t, tc.expected, intent)
        })
    }
}

func TestKeywordClassifier_GeneralInquiry(t *testing.T) {
    classifier := NewKeywordClassifier()
    
    testCases := []string{
        "What are your business hours?",
        "Do you ship internationally?",
        "Can I change my payment method?",
        "Tell me about your loyalty program",
    }
    
    for _, message := range testCases {
        t.Run(message, func(t *testing.T) {
            intent, _, _ := classifier.Classify(context.Background(), message, nil)
            assert.Equal(t, "general_inquiry", intent)
        })
    }
}
```

#### 1.4 Refund Amount Calculation

```go
func TestCalculateRefundAmount_FullOrder(t *testing.T) {
    // Given: Returning all items
    order := &DetailedOrder{
        Total:        7475, // R74.75 in cents
        Discount:     500,
        ShippingCost: 0,
        Parcels: []Parcel{
            {
                OrderItems: []OrderItem{
                    {SKU: "TEST001", Quantity: 1, Price: 7475},
                },
            },
        },
    }
    returnItems := []ReturnItem{
        {ItemID: "TEST001", Quantity: 1, Price: 7475},
    }
    
    // When: Calculating refund
    refundAmount := CalculateRefundAmount(order, returnItems)
    
    // Then: Should refund full order amount
    assert.Equal(t, int64(7475), refundAmount)
}

func TestCalculateRefundAmount_PartialReturn(t *testing.T) {
    // Given: Returning 1 of 2 items
    order := &DetailedOrder{
        Total: 10000,
        Parcels: []Parcel{
            {
                OrderItems: []OrderItem{
                    {SKU: "TEST001", Quantity: 1, Price: 6000},
                    {SKU: "TEST002", Quantity: 1, Price: 4000},
                },
            },
        },
    }
    returnItems := []ReturnItem{
        {ItemID: "TEST001", Quantity: 1, Price: 6000},
    }
    
    // When: Calculating refund
    refundAmount := CalculateRefundAmount(order, returnItems)
    
    // Then: Should refund only returned item price
    assert.Equal(t, int64(6000), refundAmount)
}

func TestCalculateRefundAmount_MultipleQuantities(t *testing.T) {
    // Given: Returning 2 of 3 identical items
    order := &DetailedOrder{
        Parcels: []Parcel{
            {
                OrderItems: []OrderItem{
                    {SKU: "TEST001", Quantity: 3, Price: 2000}, // R20 each
                },
            },
        },
    }
    returnItems := []ReturnItem{
        {ItemID: "TEST001", Quantity: 2, Price: 2000},
    }
    
    // When: Calculating refund
    refundAmount := CalculateRefundAmount(order, returnItems)
    
    // Then: Should refund 2 items
    assert.Equal(t, int64(4000), refundAmount) // 2 * 2000
}
```

### 2. Integration Tests - API Endpoints

#### 2.1 Conversation Management

```go
package integration_test

func TestCreateConversation_NaturalLanguage_Success(t *testing.T) {
    db := setupTestDB(t)
    defer db.Close()
    router := setupTestRouter(db)
    
    // Given: Valid conversation request
    reqBody := `{
        "type": "natural_language",
        "initial_message": "I want to return my Nike shoes"
    }`
    
    // When: Creating conversation
    w := performRequest(router, "POST", "/api/v1/conversations", reqBody, validAuthToken())
    
    // Then: Should create successfully
    assert.Equal(t, http.StatusCreated, w.Code)
    
    var response map[string]interface{}
    json.Unmarshal(w.Body.Bytes(), &response)
    
    assert.NotEmpty(t, response["conversation_id"])
    assert.Equal(t, "natural_language", response["type"])
    assert.Equal(t, "return", response["intent"])
    assert.NotEmpty(t, response["message"])
}

func TestCreateConversation_GuidedFlow_Success(t *testing.T) {
    db := setupTestDB(t)
    defer db.Close()
    router := setupTestRouter(db)
    
    // Given: Guided flow request
    reqBody := `{
        "type": "guided_flow",
        "intent": "return"
    }`
    
    // When: Creating conversation
    w := performRequest(router, "POST", "/api/v1/conversations", reqBody, validAuthToken())
    
    // Then: Should create successfully
    assert.Equal(t, http.StatusCreated, w.Code)
    
    var response map[string]interface{}
    json.Unmarshal(w.Body.Bytes(), &response)
    
    assert.NotEmpty(t, response["conversation_id"])
    assert.Equal(t, "guided_flow", response["type"])
    assert.NotEmpty(t, response["current_step"])
    assert.NotEmpty(t, response["options"])
}

func TestGetConversation_Success(t *testing.T) {
    db := setupTestDB(t)
    defer db.Close()
    router := setupTestRouter(db)
    
    // Given: Existing conversation
    conversationID := createTestConversation(db, "test-customer-id")
    
    // When: Fetching conversation
    w := performRequest(router, "GET", "/api/v1/conversations/"+conversationID, "", validAuthToken())
    
    // Then: Should return conversation
    assert.Equal(t, http.StatusOK, w.Code)
    
    var response map[string]interface{}
    json.Unmarshal(w.Body.Bytes(), &response)
    
    assert.Equal(t, conversationID, response["conversation_id"])
    assert.NotEmpty(t, response["messages"])
}

func TestGetConversation_NotFound(t *testing.T) {
    db := setupTestDB(t)
    defer db.Close()
    router := setupTestRouter(db)
    
    // Given: Non-existent conversation ID
    fakeID := uuid.New().String()
    
    // When: Fetching conversation
    w := performRequest(router, "GET", "/api/v1/conversations/"+fakeID, "", validAuthToken())
    
    // Then: Should return 404
    assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGetConversation_UnauthorizedAccess(t *testing.T) {
    db := setupTestDB(t)
    defer db.Close()
    router := setupTestRouter(db)
    
    // Given: Conversation belonging to different customer
    conversationID := createTestConversation(db, "customer-1")
    
    // When: Different customer tries to access
    otherCustomerToken := generateTestJWT("customer-2", "other@example.com")
    w := performRequest(router, "GET", "/api/v1/conversations/"+conversationID, "", otherCustomerToken)
    
    // Then: Should return 403
    assert.Equal(t, http.StatusForbidden, w.Code)
}
```

#### 2.2 Message Processing

```go
func TestSendMessage_SuccessfulIntent(t *testing.T) {
    db := setupTestDB(t)
    defer db.Close()
    router := setupTestRouter(db)
    
    // Given: Existing conversation
    conversationID := createTestConversation(db, "test-customer-id")
    
    reqBody := `{
        "content": "I want to return order B22881153-01"
    }`
    
    // When: Sending message
    w := performRequest(router, "POST", 
        "/api/v1/conversations/"+conversationID+"/messages", 
        reqBody, validAuthToken())
    
    // Then: Should process successfully
    assert.Equal(t, http.StatusOK, w.Code)
    
    var response map[string]interface{}
    json.Unmarshal(w.Body.Bytes(), &response)
    
    message := response["message"].(map[string]interface{})
    assert.Equal(t, "assistant", message["role"])
    assert.NotEmpty(t, message["content"])
}

func TestSendMessage_EmptyContent(t *testing.T) {
    db := setupTestDB(t)
    defer db.Close()
    router := setupTestRouter(db)
    
    conversationID := createTestConversation(db, "test-customer-id")
    
    reqBody := `{
        "content": ""
    }`
    
    // When: Sending empty message
    w := performRequest(router, "POST", 
        "/api/v1/conversations/"+conversationID+"/messages", 
        reqBody, validAuthToken())
    
    // Then: Should return validation error
    assert.Equal(t, http.StatusBadRequest, w.Code)
    
    var response map[string]interface{}
    json.Unmarshal(w.Body.Bytes(), &response)
    assert.Contains(t, response["error"], "content")
}
```

#### 2.3 Return Request Creation

```go
func TestCreateReturn_Success(t *testing.T) {
    db := setupTestDB(t)
    defer db.Close()
    router := setupTestRouter(db)
    ordersAPIMock := setupMockOrdersAPI(t)
    defer ordersAPIMock.Close()
    
    // Given: Valid return request
    reqBody := `{
        "conversation_id": "` + createTestConversation(db, "test-customer-id") + `",
        "order_number": "B22881153-01",
        "items": [
            {
                "item_id": "TEST001",
                "name": "Test Product",
                "quantity": 1,
                "price_cents": 7475
            }
        ],
        "reason": "wrong_size",
        "detailed_reason": "Shoes are too small",
        "refund_method": "original_payment",
        "delivery_method": "courier",
        "shipping_address": {
            "street": "123 Test St",
            "city": "Cape Town",
            "province": "Western Cape",
            "postal_code": "8001",
            "country": "ZA"
        }
    }`
    
    // When: Creating return
    w := performRequest(router, "POST", "/api/v1/returns", reqBody, validAuthToken())
    
    // Then: Should create successfully
    assert.Equal(t, http.StatusCreated, w.Code)
    
    var response map[string]interface{}
    json.Unmarshal(w.Body.Bytes(), &response)
    
    assert.NotEmpty(t, response["id"])
    assert.NotEmpty(t, response["return_reference"])
    assert.Equal(t, "pending_approval", response["status"])
    assert.Equal(t, float64(7475), response["refund_amount_cents"])
}

func TestCreateReturn_InvalidOrderNumber(t *testing.T) {
    db := setupTestDB(t)
    defer db.Close()
    router := setupTestRouter(db)
    
    reqBody := `{
        "order_number": "INVALID",
        "items": [{"item_id": "TEST001", "quantity": 1, "price_cents": 1000}],
        "reason": "wrong_size",
        "refund_method": "original_payment",
        "delivery_method": "courier"
    }`
    
    // When: Creating return with invalid order
    w := performRequest(router, "POST", "/api/v1/returns", reqBody, validAuthToken())
    
    // Then: Should return error
    assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateReturn_MissingRequiredFields(t *testing.T) {
    db := setupTestDB(t)
    defer db.Close()
    router := setupTestRouter(db)
    
    testCases := []struct {
        name     string
        reqBody  string
        missingField string
    }{
        {
            name:     "Missing items",
            reqBody:  `{"order_number": "B22881153-01", "reason": "wrong_size"}`,
            missingField: "items",
        },
        {
            name:     "Missing reason",
            reqBody:  `{"order_number": "B22881153-01", "items": []}`,
            missingField: "reason",
        },
        {
            name:     "Missing refund_method",
            reqBody:  `{"order_number": "B22881153-01", "items": [], "reason": "wrong_size"}`,
            missingField: "refund_method",
        },
    }
    
    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            w := performRequest(router, "POST", "/api/v1/returns", tc.reqBody, validAuthToken())
            assert.Equal(t, http.StatusBadRequest, w.Code)
            
            var response map[string]interface{}
            json.Unmarshal(w.Body.Bytes(), &response)
            assert.Contains(t, response["error"], tc.missingField)
        })
    }
}
```

#### 2.4 Orders API Integration

```go
func TestGetOrderHistory_Success(t *testing.T) {
    db := setupTestDB(t)
    defer db.Close()
    router := setupTestRouter(db)
    ordersAPIMock := setupMockOrdersAPI(t)
    defer ordersAPIMock.Close()
    
    // When: Fetching order history
    w := performRequest(router, "GET", "/api/v1/orders?page=0&page_size=20", "", validAuthToken())
    
    // Then: Should return orders
    assert.Equal(t, http.StatusOK, w.Code)
    
    var response map[string]interface{}
    json.Unmarshal(w.Body.Bytes(), &response)
    
    orders := response["orders"].([]interface{})
    assert.Greater(t, len(orders), 0)
    
    firstOrder := orders[0].(map[string]interface{})
    assert.NotEmpty(t, firstOrder["orderNumber"])
    assert.NotEmpty(t, firstOrder["orderDate"])
}

func TestGetOrderDetail_Success(t *testing.T) {
    db := setupTestDB(t)
    defer db.Close()
    router := setupTestRouter(db)
    ordersAPIMock := setupMockOrdersAPI(t)
    defer ordersAPIMock.Close()
    
    // When: Fetching order details
    w := performRequest(router, "GET", "/api/v1/orders/B22881153-01", "", validAuthToken())
    
    // Then: Should return order details with parcels
    assert.Equal(t, http.StatusOK, w.Code)
    
    var response map[string]interface{}
    json.Unmarshal(w.Body.Bytes(), &response)
    
    assert.Equal(t, "B22881153-01", response["orderNumber"])
    assert.NotEmpty(t, response["parcels"])
    
    parcels := response["parcels"].([]interface{})
    assert.Greater(t, len(parcels), 0)
    
    firstParcel := parcels[0].(map[string]interface{})
    assert.NotEmpty(t, firstParcel["orderItems"])
}

func TestGetOrderDetail_NotOwned(t *testing.T) {
    db := setupTestDB(t)
    defer db.Close()
    router := setupTestRouter(db)
    ordersAPIMock := setupMockOrdersAPIWith403(t)
    defer ordersAPIMock.Close()
    
    // When: Trying to access order not owned by customer
    w := performRequest(router, "GET", "/api/v1/orders/B99999999-01", "", validAuthToken())
    
    // Then: Should return 403
    assert.Equal(t, http.StatusForbidden, w.Code)
}
```

### 3. Error Handling Tests

#### 3.1 Authentication Errors

```go
func TestAuthMiddleware_MissingToken(t *testing.T) {
    db := setupTestDB(t)
    defer db.Close()
    router := setupTestRouter(db)
    
    // When: Request without Authorization header
    w := performRequest(router, "GET", "/api/v1/conversations", "", "")
    
    // Then: Should return 401
    assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
    db := setupTestDB(t)
    defer db.Close()
    router := setupTestRouter(db)
    
    // When: Request with malformed token
    w := performRequest(router, "GET", "/api/v1/conversations", "", "Bearer invalid-token")
    
    // Then: Should return 401
    assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthMiddleware_ExpiredToken(t *testing.T) {
    db := setupTestDB(t)
    defer db.Close()
    router := setupTestRouter(db)
    
    // Given: Expired JWT
    expiredToken := generateExpiredJWT("test-customer", "test@example.com")
    
    // When: Request with expired token
    w := performRequest(router, "GET", "/api/v1/conversations", "", expiredToken)
    
    // Then: Should return 401
    assert.Equal(t, http.StatusUnauthorized, w.Code)
}
```

#### 3.2 Validation Errors

```go
func TestValidation_InvalidConversationType(t *testing.T) {
    db := setupTestDB(t)
    defer db.Close()
    router := setupTestRouter(db)
    
    reqBody := `{
        "type": "invalid_type",
        "initial_message": "test"
    }`
    
    // When: Creating conversation with invalid type
    w := performRequest(router, "POST", "/api/v1/conversations", reqBody, validAuthToken())
    
    // Then: Should return 400
    assert.Equal(t, http.StatusBadRequest, w.Code)
    
    var response map[string]interface{}
    json.Unmarshal(w.Body.Bytes(), &response)
    assert.Contains(t, response["error"], "type")
}

func TestValidation_InvalidRefundMethod(t *testing.T) {
    db := setupTestDB(t)
    defer db.Close()
    router := setupTestRouter(db)
    
    reqBody := `{
        "order_number": "B22881153-01",
        "items": [{"item_id": "TEST001", "quantity": 1, "price_cents": 1000}],
        "reason": "wrong_size",
        "refund_method": "invalid_method",
        "delivery_method": "courier"
    }`
    
    // When: Creating return with invalid refund method
    w := performRequest(router, "POST", "/api/v1/returns", reqBody, validAuthToken())
    
    // Then: Should return 400
    assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestValidation_NegativeQuantity(t *testing.T) {
    db := setupTestDB(t)
    defer db.Close()
    router := setupTestRouter(db)
    
    reqBody := `{
        "order_number": "B22881153-01",
        "items": [{"item_id": "TEST001", "quantity": -1, "price_cents": 1000}],
        "reason": "wrong_size",
        "refund_method": "original_payment",
        "delivery_method": "courier"
    }`
    
    // When: Creating return with negative quantity
    w := performRequest(router, "POST", "/api/v1/returns", reqBody, validAuthToken())
    
    // Then: Should return 400
    assert.Equal(t, http.StatusBadRequest, w.Code)
}
```

#### 3.3 Resource Not Found

```go
func TestNotFound_Conversation(t *testing.T) {
    db := setupTestDB(t)
    defer db.Close()
    router := setupTestRouter(db)
    
    fakeID := uuid.New().String()
    
    // When: Accessing non-existent conversation
    w := performRequest(router, "GET", "/api/v1/conversations/"+fakeID, "", validAuthToken())
    
    // Then: Should return 404
    assert.Equal(t, http.StatusNotFound, w.Code)
    
    var response map[string]interface{}
    json.Unmarshal(w.Body.Bytes(), &response)
    assert.Contains(t, response["error"], "not found")
}

func TestNotFound_Return(t *testing.T) {
    db := setupTestDB(t)
    defer db.Close()
    router := setupTestRouter(db)
    
    fakeID := uuid.New().String()
    
    // When: Accessing non-existent return
    w := performRequest(router, "GET", "/api/v1/returns/"+fakeID, "", validAuthToken())
    
    // Then: Should return 404
    assert.Equal(t, http.StatusNotFound, w.Code)
}
```

### 4. SQLite-Specific Tests

#### 4.1 JSON Serialization/Deserialization

```go
package persistence_test

func TestSQLite_JSONContextSerialization(t *testing.T) {
    db := setupSQLiteDB(t)
    defer db.Close()
    
    // Given: Conversation with complex context
    conversation := &entities.Conversation{
        ID:         uuid.New().String(),
        CustomerID: "test-customer",
        Type:       "natural_language",
        Intent:     "return",
        Context: map[string]any{
            "order_number": "B22881153-01",
            "selected_items": []string{"TEST001", "TEST002"},
            "nested_data": map[string]any{
                "key1": "value1",
                "key2": 123,
            },
        },
    }
    
    // When: Saving to SQLite
    repo := NewConversationRepository(db)
    err := repo.Create(context.Background(), conversation)
    assert.NoError(t, err)
    
    // Then: Should retrieve with correct context
    retrieved, err := repo.GetByID(context.Background(), conversation.ID)
    assert.NoError(t, err)
    assert.Equal(t, conversation.Context["order_number"], retrieved.Context["order_number"])
    assert.Len(t, retrieved.Context["selected_items"].([]interface{}), 2)
}

func TestSQLite_JSONArrayInItems(t *testing.T) {
    db := setupSQLiteDB(t)
    defer db.Close()
    
    // Given: Return request with items array
    returnReq := &entities.ReturnRequest{
        ID:         uuid.New().String(),
        CustomerID: "test-customer",
        OrderNumber: "B22881153-01",
        Items: []entities.ReturnItem{
            {ItemID: "TEST001", Name: "Item 1", Quantity: 1, Price: 5000},
            {ItemID: "TEST002", Name: "Item 2", Quantity: 2, Price: 3000},
        },
        Reason:             "wrong_size",
        RefundMethod:       "original_payment",
        DeliveryMethod:     "courier",
        RefundAmountCents:  11000,
    }
    
    // When: Saving to SQLite
    repo := NewReturnRepository(db)
    err := repo.Create(context.Background(), returnReq)
    assert.NoError(t, err)
    
    // Then: Should retrieve with correct items
    retrieved, err := repo.GetByID(context.Background(), returnReq.ID)
    assert.NoError(t, err)
    assert.Len(t, retrieved.Items, 2)
    assert.Equal(t, "TEST001", retrieved.Items[0].ItemID)
    assert.Equal(t, 2, retrieved.Items[1].Quantity)
}
```

#### 4.2 UUID String Handling

```go
func TestSQLite_UUIDStorage(t *testing.T) {
    db := setupSQLiteDB(t)
    defer db.Close()
    
    // Given: Multiple entities with UUIDs
    conv1ID := uuid.New().String()
    conv2ID := uuid.New().String()
    
    conversation1 := createTestConversationWithID(conv1ID, "customer-1")
    conversation2 := createTestConversationWithID(conv2ID, "customer-1")
    
    repo := NewConversationRepository(db)
    repo.Create(context.Background(), conversation1)
    repo.Create(context.Background(), conversation2)
    
    // When: Retrieving by UUID string
    retrieved1, err := repo.GetByID(context.Background(), conv1ID)
    assert.NoError(t, err)
    
    retrieved2, err := repo.GetByID(context.Background(), conv2ID)
    assert.NoError(t, err)
    
    // Then: Should retrieve correct conversations
    assert.Equal(t, conv1ID, retrieved1.ID)
    assert.Equal(t, conv2ID, retrieved2.ID)
    assert.NotEqual(t, retrieved1.ID, retrieved2.ID)
}
```

#### 4.3 DateTime Format Handling

```go
func TestSQLite_DateTimeFormatting(t *testing.T) {
    db := setupSQLiteDB(t)
    defer db.Close()
    
    // Given: Conversation with specific timestamp
    now := time.Now()
    conversation := &entities.Conversation{
        ID:         uuid.New().String(),
        CustomerID: "test-customer",
        CreatedAt:  now,
        UpdatedAt:  now,
    }
    
    // When: Saving and retrieving
    repo := NewConversationRepository(db)
    err := repo.Create(context.Background(), conversation)
    assert.NoError(t, err)
    
    retrieved, err := repo.GetByID(context.Background(), conversation.ID)
    assert.NoError(t, err)
    
    // Then: Timestamps should be preserved (within 1 second tolerance)
    assert.WithinDuration(t, conversation.CreatedAt, retrieved.CreatedAt, time.Second)
    assert.WithinDuration(t, conversation.UpdatedAt, retrieved.UpdatedAt, time.Second)
}
```

#### 4.4 Boolean Integer Conversion

```go
func TestSQLite_BooleanAsInteger(t *testing.T) {
    db := setupSQLiteDB(t)
    defer db.Close()
    
    // Given: Flow with is_active boolean
    activeFlow := &entities.Flow{
        ID:       uuid.New().String(),
        Name:     "Active Flow",
        Intent:   "return",
        IsActive: true,
        Steps:    []entities.FlowStep{},
    }
    
    inactiveFlow := &entities.Flow{
        ID:       uuid.New().String(),
        Name:     "Inactive Flow",
        Intent:   "refund",
        IsActive: false,
        Steps:    []entities.FlowStep{},
    }
    
    // When: Saving both flows
    repo := NewFlowRepository(db)
    repo.Create(context.Background(), activeFlow)
    repo.Create(context.Background(), inactiveFlow)
    
    // Then: Should retrieve with correct boolean values
    retrieved1, _ := repo.GetByID(context.Background(), activeFlow.ID)
    assert.True(t, retrieved1.IsActive)
    
    retrieved2, _ := repo.GetByID(context.Background(), inactiveFlow.ID)
    assert.False(t, retrieved2.IsActive)
    
    // And: Should filter by boolean correctly
    activeFlows, _ := repo.GetActiveFlows(context.Background())
    assert.Len(t, activeFlows, 1)
    assert.Equal(t, "Active Flow", activeFlows[0].Name)
}
```

### 5. Test Helpers & Utilities

```go
// tests/helpers/setup.go
package helpers

func SetupSQLiteDB(t *testing.T) *sql.DB {
    // Create in-memory SQLite database for testing
    db, err := sql.Open("sqlite", ":memory:")
    if err != nil {
        t.Fatalf("Failed to create test database: %v", err)
    }
    
    // Run migrations
    if err := runMigrations(db, "../../migrations"); err != nil {
        t.Fatalf("Failed to run migrations: %v", err)
    }
    
    // Seed initial data
    if err := seedTestData(db); err != nil {
        t.Fatalf("Failed to seed test data: %v", err)
    }
    
    return db
}

func SetupMockOrdersAPI(t *testing.T) *httptest.Server {
    return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        switch {
        case strings.HasSuffix(r.URL.Path, "/orders/history"):
            fixture := LoadFixture("order_history_response.json")
            w.Header().Set("Content-Type", "application/json")
            w.Write(fixture)
            
        case strings.Contains(r.URL.Path, "/order/B22881153-01"):
            fixture := LoadFixture("order_detail_response.json")
            w.Header().Set("Content-Type", "application/json")
            w.Write(fixture)
            
        default:
            w.WriteHeader(http.StatusNotFound)
        }
    }))
}

func SetupMockOrdersAPIWith403(t *testing.T) *httptest.Server {
    return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusForbidden)
    }))
}

func PerformRequest(router *gin.Engine, method, path, body, authToken string) *httptest.ResponseRecorder {
    req := httptest.NewRequest(method, path, strings.NewReader(body))
    
    if authToken != "" {
        req.Header.Set("Authorization", authToken)
    }
    
    if body != "" {
        req.Header.Set("Content-Type", "application/json")
    }
    
    w := httptest.NewRecorder()
    router.ServeHTTP(w, req)
    
    return w
}

func ValidAuthToken() string {
    return GenerateTestJWT("test-customer-id", "test@example.com")
}

func GenerateExpiredJWT(customerID, email string) string {
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
        "sub":   customerID,
        "email": email,
        "exp":   time.Now().Add(-24 * time.Hour).Unix(), // Expired yesterday
    })
    
    tokenString, _ := token.SignedString([]byte("test-secret"))
    return "Bearer " + tokenString
}

func CreateTestConversation(db *sql.DB, customerID string) string {
    id := uuid.New().String()
    _, err := db.Exec(`
        INSERT INTO conversations (id, customer_id, type, intent, state, created_at, updated_at)
        VALUES (?, ?, ?, ?, ?, ?, ?)
    `, id, customerID, "natural_language", "return", "active", 
       time.Now().Format(time.RFC3339), time.Now().Format(time.RFC3339))
    
    if err != nil {
        panic(err)
    }
    
    return id
}
```

### 6. Test Execution & Coverage

#### Running Tests

```bash
# Run all tests
go test ./... -v

# Run with coverage
go test ./... -cover -coverprofile=coverage.out

# View coverage report
go tool cover -html=coverage.out

# Run specific test suite
go test ./tests/unit -v
go test ./tests/integration -v
go test ./tests/e2e -v

# Run tests with race detection
go test ./... -race

# Run tests matching pattern
go test ./... -run TestOrderItem

# Run with verbose output and no cache
go test ./... -v -count=1
```

#### Coverage Goals

- **Unit Tests**: >80% code coverage
- **Integration Tests**: All API endpoints covered
- **E2E Tests**: Critical user journeys (Returns flow, Order tracking)
- **Error Handling**: All error paths tested

#### CI/CD Integration

```yaml
# .github/workflows/test.yml
name: Test Suite

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      
      - name: Install dependencies
        run: go mod download
      
      - name: Run unit tests
        run: go test ./internal/... -v -cover
      
      - name: Run integration tests
        run: go test ./tests/integration/... -v
      
      - name: Run E2E tests
        run: go test ./tests/e2e/... -v
      
      - name: Check coverage
        run: |
          go test ./... -coverprofile=coverage.out
          go tool cover -func=coverage.out
```

## Monitoring & Observability

### Prometheus Metrics

```go
package metrics

import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    ConversationsCreated = promauto.NewCounter(prometheus.CounterOpts{
        Name: "chatbot_conversations_created_total",
        Help: "Total number of conversations created",
    })

    MessagesProcessed = promauto.NewCounterVec(prometheus.CounterOpts{
        Name: "chatbot_messages_processed_total",
        Help: "Total number of messages processed",
    }, []string{"intent"})

    ReturnsCreated = promauto.NewCounter(prometheus.CounterOpts{
        Name: "chatbot_returns_created_total",
        Help: "Total number of returns created",
    })

    AILatency = promauto.NewHistogram(prometheus.HistogramOpts{
        Name:    "chatbot_ai_latency_seconds",
        Help:    "AI processing latency in seconds",
        Buckets: prometheus.DefBuckets,
    })

    OrderAPILatency = promauto.NewHistogram(prometheus.HistogramOpts{
        Name:    "chatbot_orders_api_latency_seconds",
        Help:    "Orders API latency in seconds",
        Buckets: prometheus.DefBuckets,
    })
)
```

## Development Roadmap & Phasing

### Phase 1: Core Returns Flow (Weeks 1-6) - **PRIMARY FOCUS**

**Week 1-2: Foundation**
- [ ] Project setup (Go modules, directory structure)
- [ ] Database schema implementation
- [ ] Core domain entities (Conversation, ReturnRequest, Flow)
- [ ] JWT authentication middleware
- [ ] Basic HTTP handlers and routing

**Week 3-4: Returns Implementation**
- [ ] Orders API client integration (read-only)
- [ ] Natural language intent classification (OpenAI/Anthropic)
- [ ] Return flow logic (identify order → select items → reason → confirm)
- [ ] Conversation state management
- [ ] Redis session caching

**Week 5-6: Returns Completion**
- [ ] Return Management API integration (if available)
- [ ] Photo upload to S3/storage
- [ ] Email notifications
- [ ] Flow configuration seeding (database migrations)
- [ ] Comprehensive testing

**Deliverables:**
- Working returns flow (natural language + guided)
- Order history integration
- Basic conversation management
- Internal flow management API

### Phase 2: Additional Flow Types (Weeks 7-10)

**Week 7-8: Refund Status & Order Tracking**
- [ ] Refund status tracking endpoints
- [ ] Refund timeline display
- [ ] Enhanced order tracking with parcels
- [ ] SMS notification opt-in

**Week 9-10: Account Management Flows**
- [ ] Password reset flow
- [ ] OTP generation and verification
- [ ] Profile update flows
- [ ] Address change functionality

**Deliverables:**
- Refund status checking
- Enhanced order tracking
- Account management basics
- Multi-flow intent routing

### Phase 3: Scale & Optimize (Weeks 11-12)

**Week 11: Performance & Reliability**
- [ ] Load testing and optimization
- [ ] AI response caching
- [ ] Database query optimization
- [ ] Horizontal scaling validation

**Week 12: Monitoring & Analytics**
- [ ] Prometheus metrics dashboard
- [ ] Conversation analytics
- [ ] Intent classification accuracy tracking
- [ ] A/B testing framework (for flow variations)

**Deliverables:**
- Production-ready API
- Monitoring infrastructure
- Performance benchmarks
- Analytics dashboard

### Future Phases (Post-Launch)

**Phase 4: Advanced Features**
- [ ] WebSocket for real-time chat
- [ ] Multi-language support (i18n)
- [ ] Voice input/output (speech-to-text)
- [ ] Sentiment analysis
- [ ] Proactive support suggestions
- [ ] Agent escalation/handoff

**Phase 5: Admin Interface** (Out of current scope)
- [ ] Web-based flow builder UI
- [ ] Visual flow editor
- [ ] Flow versioning UI
- [ ] Analytics dashboard
- [ ] A/B test configuration UI

**Note**: Admin interface is explicitly out of scope for the initial API development. Flow management will be handled via:
1. Database migrations (initial flows)
2. Internal API endpoints (programmatic updates)
3. Future admin UI (separate project)

## Multi-Flow Architecture Considerations

### Intent Routing Strategy

The system uses a layered approach for handling multiple intents:

1. **AI Classification** (Natural Language)
   - Primary: OpenAI/Anthropic for intent detection
   - Fallback: Keyword matching if AI fails
   - Confidence threshold: 0.80 for primary intents

2. **Explicit Selection** (Guided Flows)
   - User selects from menu
   - No ambiguity in intent

3. **Context Awareness**
   - Previous conversation history informs classification
   - Multi-turn conversations maintain context
   - State machine for flow progression

### Flow Configuration Pattern

Each flow follows a consistent pattern:

```go
type FlowDefinition struct {
    Intent string
    Steps []Step
    EntryConditions []Condition // e.g., "order must be eligible for return"
    ExitActions []Action        // e.g., "send confirmation email"
}
```

**Example Flow Definitions:**

1. **Returns Flow** - Multi-step with branching
2. **Track Order Flow** - Linear, data display focused
3. **Refund Status Flow** - Query-based, read-only
4. **Password Reset Flow** - Security-focused, OTP verification
5. **Address Update Flow** - Data modification, validation heavy

### Extensibility Principles

To add a new flow type:

1. **Define Intent** - Add to AI classifier prompt
2. **Create Entity** - New domain entity if needed (e.g., `CancellationRequest`)
3. **Implement Handler** - Add case to intent switch in `ProcessMessageUseCase`
4. **Configure Flow** - Seed flow configuration in database
5. **Add Tests** - Unit and integration tests for new flow

The architecture is designed to make adding new flows straightforward without modifying core conversation logic.

## Security Checklist

- [ ] JWT token validation on all endpoints
- [ ] Order ownership verification
- [ ] Rate limiting (per customer)
- [ ] Input sanitization and validation
- [ ] SQL injection prevention (parameterized queries)
- [ ] CORS configuration
- [ ] Secrets management (AWS Secrets Manager)
- [ ] TLS/HTTPS enforcement
- [ ] Audit logging for all return operations
- [ ] PII data encryption at rest

## Performance Considerations

- **Caching**: Cache order data in Redis (TTL: 5 minutes)
- **AI Rate Limiting**: Implement exponential backoff for AI API
- **Database Connection Pooling**: Max 25 connections
- **Horizontal Scaling**: Stateless API design for easy scaling
- **Async Processing**: Queue for photo uploads and email notifications
- **Response Compression**: Gzip compression for API responses
- **Intent Classification Cache**: Cache common queries to reduce AI API calls
- **Flow Configuration Cache**: Redis cache for active flows (invalidate on updates)

## Summary

### What's In Scope

**Phase 1 (Primary Focus):**
- Returns flow (natural language + guided)
- Order integration (history, details, tracking)
- Conversation management
- JWT authentication
- AI-powered intent classification
- Basic flow configuration via database/API

**Phase 2 (Future):**
- Refund status tracking
- Account management (password reset, profile updates)
- Enhanced order tracking
- Additional support flows

### What's Out of Scope

- **Admin Web UI** - Flow management via API only (UI is separate future project)
- **Human Agent Handoff** - Escalation mentioned but not implemented
- **Voice/Speech** - Text-based only initially
- **Multi-language** - English only in Phase 1
- **Video Chat** - Not planned
- **Payment Processing** - Integrated with existing systems only

### Key Design Decisions

1. **API-First Approach** - No admin UI initially, all configuration via API endpoints and database seeding
2. **Clean Architecture** - Clear separation of concerns (domain, application, infrastructure, interfaces)
3. **Multi-Flow from Day 1** - Architecture supports multiple intents, even if only Returns is fully implemented initially
4. **AI-Powered NL** - OpenAI/Anthropic for natural language understanding, not rule-based
5. **Stateful Conversations** - PostgreSQL stores full conversation history for context
6. **Extensible Flow System** - JSON-based flow configurations stored in database
7. **Hybrid Interaction** - Support both natural language and structured guided flows

### Success Metrics

- **Customer Satisfaction**: >80% successful return completions without human intervention
- **Response Time**: <2s average for AI classification and response generation
- **Intent Accuracy**: >90% correct intent classification
- **Availability**: 99.9% uptime SLA
- **Scalability**: Handle 1000+ concurrent conversations

### Next Steps

1. **Review & Approval** - Stakeholder sign-off on specification
2. **Environment Setup** - Provision infrastructure (PostgreSQL, Redis, S3, AI API keys)
3. **Sprint Planning** - Break down Phase 1 into 2-week sprints
4. **Development Kickoff** - Begin with project scaffold and core entities
5. **Integration Coordination** - Align with Orders API team on integration patterns

## MVP/POC Implementation Checklist

This checklist provides a step-by-step guide to implement a working MVP with E2E tests.

### Week 1: Project Setup & Foundation

**Day 1-2: Project Scaffolding**
- [ ] Initialize Go module (`go mod init customer-support-api`)
- [ ] Create directory structure as per specification
- [ ] Setup `.gitignore` (vendor/, .env, *.log)
- [ ] Create `Makefile` with common commands (build, test, run, migrate)
- [ ] Setup Docker Compose for local development (PostgreSQL, Redis)
- [ ] Create basic `main.go` with HTTP server

**Day 3: Database & Migrations**
- [ ] Install `golang-migrate` or similar migration tool
- [ ] Create migration files from schema specification
- [ ] Add `make migrate-up` and `make migrate-down` commands
- [ ] Create seed script for initial flow data
- [ ] Test migrations run successfully

**Day 4-5: Core Domain Entities**
- [ ] Implement `Conversation` entity in `internal/domain/entities/conversation.go`
- [ ] Implement `ReturnRequest` entity
- [ ] Implement `Flow` entity
- [ ] Implement `Message` entity
- [ ] Add basic validation methods to entities
- [ ] Write unit tests for entity validation

### Week 2: API Layer & Orders Integration

**Day 1-2: Orders API Client**
- [ ] Create `OrdersClient` in `internal/infrastructure/api/orders_client.go`
- [ ] Implement `GetOrderHistory()` with response unwrapping
- [ ] Implement `GetDetailedOrder()` with parcel handling
- [ ] Create test fixtures from specification
- [ ] Write unit tests with httptest for Orders client
- [ ] Add helper functions for extracting items from parcels

**Day 3: HTTP Handlers**
- [ ] Setup Gin router in `internal/interfaces/http/router.go`
- [ ] Implement health check endpoint (`GET /health`)
- [ ] Create conversation handler skeleton
- [ ] Create return handler skeleton
- [ ] Create orders proxy handler
- [ ] Add CORS middleware

**Day 4-5: Authentication**
- [ ] Implement JWT middleware in `internal/infrastructure/middleware/auth.go`
- [ ] Add JWT generation helper for tests
- [ ] Test auth middleware with valid/invalid tokens
- [ ] Add customer ID extraction from JWT
- [ ] Create test helper `generateTestJWT()`

### Week 3: Core Use Cases

**Day 1-2: Conversation Management**
- [ ] Implement `ConversationService` in `internal/domain/services/`
- [ ] Implement `ConversationRepository` (PostgreSQL)
- [ ] Create conversation use case: `CreateConversation`
- [ ] Create use case: `GetConversation`
- [ ] Write unit tests for conversation service
- [ ] Write integration tests for conversation endpoints

**Day 3-4: Return Flow Logic**
- [ ] Implement `ReturnService`
- [ ] Implement `ReturnRepository`
- [ ] Create use case: `InitiateReturn`
- [ ] Create use case: `CreateReturnRequest`
- [ ] Add order eligibility validation
- [ ] Write tests for return logic

**Day 5: Flow Management**
- [ ] Implement `FlowService`
- [ ] Implement `FlowRepository`
- [ ] Create use case: `GetActiveFlowByIntent`
- [ ] Add flow step navigation logic
- [ ] Test flow loading from database

### Week 4: AI Integration & Testing

**Day 1-2: AI Intent Classification (Simplified for MVP)**
- [ ] Create `IntentClassifier` interface
- [ ] Implement OpenAI client OR simple keyword-based classifier for MVP
- [ ] Add intent classification to message processing
- [ ] Create mock classifier for tests
- [ ] Test intent detection accuracy

**Day 3: Message Processing**
- [ ] Implement `ProcessMessage` use case
- [ ] Add intent routing logic (return, track_order, general)
- [ ] Implement `handleReturnIntent` helper
- [ ] Implement `handleTrackOrderIntent` helper
- [ ] Test message processing flow

**Day 4-5: E2E Tests**
- [ ] Setup E2E test suite structure
- [ ] Create test database setup/teardown
- [ ] Implement `TestCompleteReturnFlow_NaturalLanguage`
- [ ] Implement `TestOrderHistoryAndTracking`
- [ ] Create mock Orders API server for tests
- [ ] Verify all tests pass

### Week 5: Polish & Documentation

**Day 1-2: Error Handling & Validation**
- [ ] Add structured error responses
- [ ] Implement request validation middleware
- [ ] Add proper HTTP status codes
- [ ] Test error scenarios (404, 401, 400, 500)
- [ ] Add logging throughout

**Day 3: Configuration**
- [ ] Implement config loading from YAML
- [ ] Add environment variable overrides
- [ ] Create `.env.example` file
- [ ] Document all configuration options
- [ ] Test with different configurations

**Day 4: Docker & Local Development**
- [ ] Create `Dockerfile` from specification
- [ ] Update `docker-compose.yml` with app service
- [ ] Test full stack runs locally
- [ ] Create `README.md` with setup instructions
- [ ] Add database initialization script

**Day 5: API Documentation**
- [ ] Generate OpenAPI/Swagger spec
- [ ] Add example requests/responses
- [ ] Test API with Postman/Insomnia
- [ ] Create basic API documentation
- [ ] Demo MVP to stakeholders

### Minimum Viable Endpoints for MVP

These endpoints are sufficient to demonstrate Returns flow:

**Must Have:**
1. `POST /api/v1/conversations` - Start conversation
2. `POST /api/v1/conversations/:id/messages` - Send message
3. `GET /api/v1/conversations/:id` - Get conversation
4. `POST /api/v1/returns` - Create return request
5. `GET /api/v1/orders` - Get order history (proxy)
6. `GET /api/v1/orders/:orderNumber` - Get order details (proxy)
7. `GET /health` - Health check

**Optional for MVP (add if time permits):**
- `POST /api/v1/conversations/:id/flow-action` - Guided flow actions
- `GET /api/v1/returns` - List returns
- `GET /api/v1/returns/:id` - Get return details

### MVP Success Criteria

The MVP is complete when:

✅ **E2E Return Flow Works**
- User can start conversation with "I want to return X"
- System fetches orders from Orders API
- User can select order and items
- User can specify return reason
- Return request is created and stored

✅ **Tests Pass**
- All unit tests pass (`make test-unit`)
- All integration tests pass (`make test-integration`)
- E2E return flow test passes (`make test-e2e`)
- Orders API integration tests pass with mock

✅ **Basic Observability**
- Health check endpoint responds
- Application logs to stdout
- Errors are logged with context

✅ **Documentation Complete**
- README with setup instructions
- API endpoints documented
- Environment variables documented
- Sample .env file provided

✅ **Deployable**
- Docker build succeeds
- Docker compose stack runs
- Migrations run successfully
- App connects to PostgreSQL and Redis

### Quick Start Commands for Development (SQLite - No Docker)

```bash
# Initial setup
cd /Users/simmbiote/Projects/bash-returns
mkdir -p data uploads migrations cmd/api internal

# Initialize Go module
go mod init customer-support-api

# Install dependencies (pure Go, no CGO required!)
go get github.com/gin-gonic/gin
go get github.com/google/uuid
go get github.com/golang-jwt/jwt/v5
go get modernc.org/sqlite  # Pure Go SQLite driver - compiles anywhere!
go get github.com/joho/godotenv

# Create .env file
cat > .env << 'EOF'
PORT=8080
DB_TYPE=sqlite
DB_PATH=./data/chatbot.db
CACHE_TYPE=memory
ORDERS_API_URL=https://web-api.bash.com
JWT_SECRET=your-test-secret-change-in-production
AI_PROVIDER=keyword
STORAGE_PROVIDER=local
STORAGE_PATH=./data/uploads
LOG_LEVEL=debug
EOF

# Create the database and run migrations
# Copy the SQLite schema from "Database Schema - SQLite Version" section above
sqlite3 ./data/chatbot.db < migrations/001_sqlite_schema.sql

# Verify tables were created
sqlite3 ./data/chatbot.db "SELECT name FROM sqlite_master WHERE type='table';"

# Create a simple main.go to test
cat > cmd/api/main.go << 'EOF'
package main

import (
    "github.com/gin-gonic/gin"
    "net/http"
)

func main() {
    r := gin.Default()
    
    r.GET("/health", func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{
            "status": "ok",
            "database": "sqlite",
            "cache": "memory",
        })
    })
    
    r.Run(":8080")
}
EOF

# Run the application
go run cmd/api/main.go

# In another terminal, test it
curl http://localhost:8080/health
# Should return: {"status":"ok","database":"sqlite","cache":"memory"}
```

### Quick Start with PostgreSQL (Production / When Docker Available)

```bash
# Create docker-compose.yml
cat > docker-compose.yml << 'EOF'
version: '3.8'
services:
  postgres:
    image: postgres:15-alpine
    environment:
      POSTGRES_DB: chatbot_db
      POSTGRES_USER: chatbot_user
      POSTGRES_PASSWORD: chatbot_pass
    ports:
      - "5432:5432"
  
  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
EOF

# Start infrastructure
docker-compose up -d

# Run PostgreSQL migrations
psql postgresql://chatbot_user:chatbot_pass@localhost:5432/chatbot_db < migrations/001_postgres_schema.sql

# Update .env
cat > .env << 'EOF'
PORT=8080
DB_TYPE=postgres
DB_HOST=localhost
DB_PORT=5432
DB_NAME=chatbot_db
DB_USER=chatbot_user
DB_PASSWORD=chatbot_pass
REDIS_HOST=localhost
REDIS_PORT=6379
ORDERS_API_URL=https://web-api.bash.com
JWT_SECRET=your-production-secret
EOF

# Run
go run cmd/api/main.go
```

### AI Integration Options for MVP

**Option 1: OpenAI Integration (Recommended)**
- More accurate intent classification
- Better natural language understanding
- Requires API key and costs per request
- ~100-200ms latency

**Option 2: Simple Keyword Matching (Faster MVP)**
- Keyword-based intent detection
- No external dependencies
- Free and instant
- Less flexible but sufficient for MVP

For POC, start with Option 2 and add Option 1 later.

## Conclusion

This specification provides a comprehensive, production-ready blueprint for a multi-flow customer support chatbot API that starts with Returns but is architected for future expansion. The API-first approach ensures flexibility while deferring admin UI complexity to a later phase.

**You now have everything needed to:**
- Set up the project structure
- Implement core functionality
- Write comprehensive tests (unit, integration, E2E)
- Deploy locally with Docker
- Demonstrate a working Returns flow

The specification includes:
✅ Complete architecture and tech stack
✅ Actual API response structures from Orders API
✅ Database schema with migrations
✅ Authentication & authorization patterns
✅ Use case implementations (pseudocode)
✅ Test fixtures and E2E test examples
✅ 5-week implementation plan
✅ Success criteria and validation steps

**Estimated MVP Timeline:** 4-5 weeks for a working POC with comprehensive tests.
