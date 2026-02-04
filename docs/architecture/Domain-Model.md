# Domain Model

## Overview

The domain model defines the core business entities and their relationships. All entities follow Domain-Driven Design principles with encapsulated state, validation logic, and behaviour methods.

## Core Entities

### Conversation

Represents a customer support interaction session, supporting both natural language and guided flow types.

**Location**: `internal/domain/entities/conversation.go`

#### Properties

| Property | Type | Required | Description |
|----------|------|----------|-------------|
| `id` | string (UUID) | Yes | Unique identifier |
| `customer_id` | string | Yes | Customer identifier |
| `type` | ConversationType | Yes | `natural_language` or `guided_flow` |
| `intent` | Intent | No | Classified intent (return, refund, track_order, etc.) |
| `current_step` | string | No | Current step in guided flow |
| `flow_id` | string | No | Flow identifier (required for guided_flow) |
| `context` | map[string]any | Yes | Key-value context store |
| `created_at` | time.Time | Yes | Creation timestamp |
| `updated_at` | time.Time | Yes | Last update timestamp |
| `completed_at` | time.Time | No | Completion timestamp |

#### Types

**ConversationType**:
- `natural_language`: Free-form conversation with AI intent classification
- `guided_flow`: Structured step-by-step flow

**Intent**:
- `return`: Customer wants to return items
- `refund`: Customer enquiring about refund
- `track_order`: Customer wants order tracking info
- `account_update`: Customer wants to update account details
- `general_inquiry`: General questions

#### Methods

**Constructor**:
```go
NewConversation(customerID string, conversationType ConversationType) *Conversation
```

**Validation**:
```go
Validate() error
```
Validates:
- ID and customer ID are present
- Type is valid (natural_language or guided_flow)
- Flow ID is present for guided flows

**Intent Management**:
```go
SetIntent(intent Intent)
```
Sets conversation intent and updates timestamp.

**Flow Management**:
```go
SetCurrentStep(step string)
```
Updates current step for guided flows.

**Context Management**:
```go
UpdateContext(key string, value any)
GetContext(key string) (any, bool)
```
Store and retrieve contextual data during conversation.

**Completion**:
```go
Complete()
IsCompleted() bool
```
Mark conversation as completed and check completion status.

#### Business Rules

1. Customer ID is mandatory for all conversations
2. Guided flows must have a flow ID
3. Context is initialised as empty map, never nil
4. Updating any field updates the `updated_at` timestamp
5. Completion sets `completed_at` to current time

---

### Message

Represents an individual message within a conversation.

**Location**: `internal/domain/entities/message.go`

#### Properties

| Property | Type | Required | Description |
|----------|------|----------|-------------|
| `id` | string (UUID) | Yes | Unique identifier |
| `conversation_id` | string | Yes | Parent conversation ID |
| `role` | MessageRole | Yes | Message sender role |
| `content` | string | Yes | Message content |
| `metadata` | map[string]any | Yes | Additional message metadata |
| `created_at` | time.Time | Yes | Creation timestamp |

#### Types

**MessageRole**:
- `user`: Message from customer
- `assistant`: Message from AI/system
- `system`: Internal system message

#### Methods

**Constructor**:
```go
NewMessage(conversationID string, role MessageRole, content string) *Message
```

**Validation**:
```go
Validate() error
```
Validates:
- ID, conversation ID, and content are present
- Role is valid (user, assistant, or system)

**Metadata Management**:
```go
AddMetadata(key string, value any)
GetMetadata(key string) (any, bool)
```
Store and retrieve message-specific metadata.

#### Business Rules

1. Messages are immutable after creation (no update methods)
2. Content cannot be empty
3. Metadata is initialised as empty map, never nil
4. Every message must belong to a conversation

---

### ReturnRequest

Represents a return request with complete workflow from draft through to refund.

**Location**: `internal/domain/entities/return_request.go`

#### Properties

| Property | Type | Required | Description |
|----------|------|----------|-------------|
| `id` | string (UUID) | Yes | Unique identifier |
| `conversation_id` | string | No | Associated conversation ID |
| `customer_id` | string | Yes | Customer identifier |
| `order_number` | string | Yes | Original order number |
| `items` | []ReturnItem | Yes | Items being returned (min 1) |
| `reason` | string | Yes | Return reason |
| `detailed_reason` | string | No | Additional reason details |
| `photos` | []string | No | Photo URLs |
| `refund_method` | RefundMethod | Yes | Refund destination |
| `delivery_method` | DeliveryMethod | Yes | Return delivery method |
| `collection_point` | string | Conditional | Required if delivery_method is collect |
| `shipping_address` | string | Conditional | Required if delivery_method is ship |
| `status` | ReturnStatus | Yes | Current status |
| `estimated_refund_date` | time.Time | No | Expected refund date |
| `refund_amount_cents` | int | Yes | Refund amount in cents |
| `return_reference` | string | No | Return tracking reference |
| `created_at` | time.Time | Yes | Creation timestamp |
| `updated_at` | time.Time | Yes | Last update timestamp |

#### Types

**ReturnStatus** (workflow):
1. `draft`: Initial state, incomplete
2. `submitted`: Customer submitted request
3. `approved`: Request approved by system/admin
4. `rejected`: Request rejected
5. `collected`: Items collected from customer
6. `received`: Items received at warehouse
7. `refunded`: Refund processed

**RefundMethod**:
- `original_payment`: Refund to original payment method
- `gift_card`: Issue as gift card
- `bash_account`: Credit to Bash account

**DeliveryMethod**:
- `collect`: Bash collects from customer
- `ship`: Customer ships to Bash

**ReturnItem**:
```go
type ReturnItem struct {
    OrderItemID string
    ProductID   string
    Quantity    int
}
```

#### Methods

**Constructor**:
```go
NewReturnRequest(customerID, orderNumber string) *ReturnRequest
```
Creates request in draft status with empty items array.

**Validation**:
```go
Validate() error
```
Validates:
- ID, customer ID, order number are present
- At least one item exists
- All items have order item ID, product ID, and positive quantity
- Reason is provided
- Refund and delivery methods are set
- Collection point set if delivery method is collect
- Shipping address set if delivery method is ship

**Item Management**:
```go
AddItem(orderItemID, productID string, quantity int) error
AddPhoto(photoURL string)
```

**Refund Management**:
```go
SetRefundAmount(amountCents int) error
```
Validates amount is non-negative.

**Status Transitions**:
```go
Submit() error                // draft → submitted
Approve(returnReference string, estimatedRefundDate time.Time)  // submitted → approved
Reject()                      // submitted → rejected
MarkCollected()               // approved → collected
MarkReceived()                // collected → received
MarkRefunded()                // received → refunded
```

#### Business Rules

1. Return requests start in draft status
2. Must have at least one item to submit
3. Submission validates entire request
4. Status transitions follow defined workflow
5. Cannot skip workflow steps
6. Refund amount must be non-negative
7. Collection point required for collect delivery
8. Shipping address required for ship delivery
9. All status transitions update `updated_at` timestamp
10. Return reference set during approval

---

## Repository Interfaces

Repositories define data access contracts without implementation details.

**Location**: `internal/domain/repositories/`

### ConversationRepository

```go
type ConversationRepository interface {
    Create(ctx context.Context, conversation *entities.Conversation) error
    GetByID(ctx context.Context, id string) (*entities.Conversation, error)
    GetByCustomerID(ctx context.Context, customerID string) ([]*entities.Conversation, error)
    Update(ctx context.Context, conversation *entities.Conversation) error
    Delete(ctx context.Context, id string) error
}
```

### MessageRepository

```go
type MessageRepository interface {
    Create(ctx context.Context, message *entities.Message) error
    GetByID(ctx context.Context, id string) (*entities.Message, error)
    GetByConversationID(ctx context.Context, conversationID string) ([]*entities.Message, error)
    Delete(ctx context.Context, id string) error
}
```

### ReturnRequestRepository

```go
type ReturnRequestRepository interface {
    Create(ctx context.Context, returnRequest *entities.ReturnRequest) error
    GetByID(ctx context.Context, id string) (*entities.ReturnRequest, error)
    GetByCustomerID(ctx context.Context, customerID string) ([]*entities.ReturnRequest, error)
    GetByOrderNumber(ctx context.Context, orderNumber string) ([]*entities.ReturnRequest, error)
    Update(ctx context.Context, returnRequest *entities.ReturnRequest) error
    Delete(ctx context.Context, id string) error
}
```

## Service Interfaces

External service contracts defined in domain layer.

**Location**: `internal/domain/services/`

### OrdersAPIClient

```go
type OrdersAPIClient interface {
    GetOrder(ctx context.Context, orderNumber string, customerID string) (*models.Order, error)
    GetOrderHistory(ctx context.Context, customerID string, page, pageSize int) (*models.OrderList, error)
}
```

Enables order validation and data retrieval for returns processing.

## Domain Models

Supporting models for external data structures.

**Location**: `internal/domain/models/`

### Order

```go
type Order struct {
    OrderNumber        string
    CustomerID         string
    Status             string
    Items              []OrderItem
    SubtotalAmountCents int
    // ... additional fields
}
```

Represents order data from Orders API.

## Entity Relationships

```
Conversation (1) ──── (N) Message
    │
    │ (optional)
    │
    └──── (0..1) ReturnRequest

Customer ──── (N) Conversation
Customer ──── (N) ReturnRequest
Order ──── (N) ReturnRequest
```

## Validation Strategy

1. **Entity-level validation**: Each entity validates its own state via `Validate()` method
2. **Constructor validation**: Basic validation during entity creation
3. **Method validation**: State transition methods validate preconditions
4. **Cross-entity validation**: Use cases validate relationships between entities

## Immutability Considerations

- **Messages**: Immutable after creation (no setter methods)
- **Conversations**: Mutable with controlled state transitions
- **Return Requests**: Mutable with strict status workflow

## Testing Coverage

Unit tests located in `internal/domain/entities/*_test.go`:
- Conversation creation and validation
- Intent and context management
- Return request workflow transitions
- Message creation and validation

**Coverage**: 70.3% (16 tests, all passing)
