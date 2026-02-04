# Returns Feature

## Overview

The returns feature enables customers to submit return requests for items from previous orders. The system handles the complete returns workflow from submission through to refund processing.

## Feature Scope

**Current Implementation (MVP)**:
- Return request creation with order validation
- Item selection and quantity specification
- Multiple refund methods (original payment, gift card, Bash account)
- Multiple delivery methods (collect, ship)
- Automatic approval for MVP (no manual review)
- Estimated refund date calculation
- Return reference generation

**Future Enhancements**:
- Manual approval workflow
- Photo upload for items
- Return status tracking
- Collection/shipping logistics integration
- Refund processing integration
- Return reason analytics

## Return Request Workflow

### Status Flow

```
draft → submitted → approved → collected → received → refunded
                  ↘ rejected
```

**Status Definitions**:

1. **draft**: Return request created but not yet submitted (incomplete data)
2. **submitted**: Customer has submitted complete return request
3. **approved**: Return request approved (automatic in MVP)
4. **rejected**: Return request rejected (not implemented in MVP)
5. **collected**: Items collected from customer (future)
6. **received**: Items received at warehouse (future)
7. **refunded**: Refund processed to customer (future)

### MVP Workflow

In the current MVP implementation:
1. Customer submits return request
2. System validates order ownership
3. System validates items can be returned
4. System calculates refund amount
5. **System automatically approves request**
6. Estimated refund date set to 7 days from submission
7. Return reference generated (format: `RET-{UUID-prefix}`)

## Use Cases

### Create Return Request

**Use Case**: `CreateReturnRequestUseCase`  
**Location**: `internal/application/usecases/create_return_request.go`

#### Input

**DTO**: `CreateReturnRequest`

```json
{
  "conversation_id": "uuid-optional",
  "order_number": "ORD-12345",
  "items": [
    {
      "order_item_id": "item-1",
      "product_id": "prod-123",
      "quantity": 1
    }
  ],
  "reason": "wrong_size",
  "detailed_reason": "The shirt is too small",
  "photos": ["https://example.com/photo1.jpg"],
  "refund_method": "original_payment",
  "delivery_method": "collect",
  "collection_point": "Store Cape Town CBD"
}
```

#### Processing Steps

1. **Order Validation**: Fetch order from Orders API
2. **Ownership Verification**: Confirm order belongs to customer
3. **Item Validation**: Verify items are returnable
4. **Refund Calculation**: Calculate refund amount from Orders API
5. **Entity Creation**: Create ReturnRequest entity
6. **Data Population**: Set all return details
7. **Validation**: Validate complete return request
8. **Submission**: Submit return (status: draft → submitted)
9. **Auto-Approval**: Approve return (status: submitted → approved)
10. **Reference Generation**: Generate return reference
11. **Persistence**: Save to database

#### Output

**DTO**: `CreateReturnResponse`

```json
{
  "id": "return-uuid",
  "return_reference": "RET-abcd1234",
  "status": "approved",
  "estimated_refund_date": "2026-02-11T10:00:00Z",
  "refund_amount_cents": 59900,
  "message": "Your return has been submitted successfully."
}
```

#### Error Scenarios

- **Order not found**: Returns validation error
- **Order ownership mismatch**: Returns invalid input error
- **Invalid items**: Items don't exist or can't be returned
- **Validation failure**: Missing required fields or invalid data
- **Database error**: Failed to persist return request

### Get Return Request

**Use Case**: `GetReturnRequestUseCase`  
**Location**: `internal/application/usecases/get_return_request.go`

Retrieves single return request by ID with ownership validation.

### List Return Requests

**Use Case**: `ListReturnRequestsUseCase`  
**Location**: `internal/application/usecases/list_return_requests.go`

Lists all return requests for a customer.

## Refund Methods

### Original Payment
**Value**: `original_payment`  
**Description**: Refund to the payment method used for the original order  
**Processing**: Integrated with payment gateway (future)

### Gift Card
**Value**: `gift_card`  
**Description**: Issue refund as a gift card/voucher  
**Processing**: Gift card generation system (future)

### Bash Account
**Value**: `bash_account`  
**Description**: Credit refund to customer's Bash account balance  
**Processing**: Account credit system (future)

## Delivery Methods

### Collect
**Value**: `collect`  
**Description**: Bash collects items from customer  
**Required Field**: `collection_point` (store location or address)  
**Process**:
1. Customer selects collection point
2. Bash arranges collection
3. Items collected from customer location
4. Status updated to "collected"

### Ship
**Value**: `ship`  
**Description**: Customer ships items back to Bash  
**Required Field**: `shipping_address` (Bash return centre address)  
**Process**:
1. Customer receives shipping label (future)
2. Customer ships items to address
3. Bash receives items at warehouse
4. Status updated to "received"

## Return Reasons

Common return reasons (stored as free text in MVP):
- `wrong_size`: Item size incorrect
- `wrong_colour`: Item colour not as expected
- `damaged`: Item arrived damaged
- `not_as_described`: Item doesn't match description
- `changed_mind`: Customer changed mind
- `quality_issue`: Quality concerns
- `arrived_late`: Delivery too late
- `ordered_by_mistake`: Accidental order

## Business Rules

### Validation Rules

1. **Order Ownership**: Customer must own the order
2. **Item Validity**: All items must exist in the original order
3. **Item Returnability**: Items must be eligible for return
4. **Quantity Limits**: Return quantity ≤ ordered quantity
5. **Required Fields**: Order number, items, reason, refund method, delivery method
6. **Conditional Fields**:
   - Collection point required if delivery method is "collect"
   - Shipping address required if delivery method is "ship"
7. **Refund Amount**: Must be non-negative

### Approval Rules (MVP)

- All valid return requests are automatically approved
- Estimated refund date set to 7 days from submission
- Return reference generated immediately

### Future Approval Rules

- Manual review required for high-value returns
- Photo evidence required for damage claims
- Automatic approval for returns within 7 days of delivery
- Rejection with reason for invalid return requests

## Integration Points

### Orders API

**Required Operations**:
1. `GetOrderByNumber()`: Fetch order details for validation
2. `ValidateOrderItems()`: Verify items are returnable
3. `CalculateRefundAmount()`: Calculate refund amount based on items

**Current State**: Mock client in MVP  
**Future**: HTTP client for production Orders API

See [Orders API Integration](../integrations/Orders-API-Integration.md) for details.

### Conversation Linking

Return requests can optionally be linked to a conversation:
- `conversation_id` field links return to originating conversation
- Enables tracking return request through chat interface
- Conversation context preserved throughout return process

## Data Model

### ReturnRequest Entity

**Location**: `internal/domain/entities/return_request.go`

Key fields:
- Order identification: `order_number`, `customer_id`
- Items: Array of `ReturnItem` (order_item_id, product_id, quantity)
- Customer input: `reason`, `detailed_reason`, `photos`
- Logistics: `refund_method`, `delivery_method`, `collection_point`, `shipping_address`
- Processing: `status`, `return_reference`, `estimated_refund_date`
- Financial: `refund_amount_cents`

### Database Persistence

**Table**: `return_requests`  
**Repository**: `ReturnRequestRepository`  
**Implementation**: `internal/infrastructure/persistence/sqlite/return_request_repository.go`

Persistence details:
- Items stored as JSON string
- Photos stored as JSON array
- Timestamps in RFC3339 format
- Nullable fields for optional data

## API Endpoints

### Create Return Request

**Endpoint**: `POST /api/v1/returns`  
**Handler**: `ReturnHandler.CreateReturn`  
**Authentication**: Customer ID required (demo mode in MVP)

**Request**:
```json
{
  "order_number": "ORD-12345",
  "items": [
    {
      "order_item_id": "item-1",
      "product_id": "prod-123",
      "quantity": 1
    }
  ],
  "reason": "wrong_size",
  "refund_method": "original_payment",
  "delivery_method": "collect",
  "collection_point": "Store Cape Town CBD"
}
```

**Response**: `201 Created`
```json
{
  "id": "return-uuid",
  "return_reference": "RET-abcd1234",
  "status": "approved",
  "estimated_refund_date": "2026-02-11T10:00:00Z",
  "refund_amount_cents": 59900,
  "message": "Your return has been submitted successfully."
}
```

### Get Return Request

**Endpoint**: `GET /api/v1/returns/:id`  
**Handler**: `ReturnHandler.GetReturn`  
**Authentication**: Customer ID required

Returns complete return request details with ownership validation.

### List Return Requests

**Endpoint**: `GET /api/v1/returns?customer_id={id}`  
**Handler**: `ReturnHandler.ListReturns`  
**Authentication**: Customer ID required

Returns array of return requests for customer.

## Testing

### Unit Tests

**Location**: `internal/domain/entities/return_request_test.go`

Tests cover:
- Return request creation
- Item addition validation
- Status transitions
- Refund amount setting
- Validation rules

### Integration Tests

Future tests will cover:
- End-to-end return request creation
- Order API integration
- Database persistence
- Status workflow transitions

## Future Enhancements

### Phase 2: Manual Approval
- Admin review interface
- Approval/rejection with reasons
- Notification system

### Phase 3: Logistics Integration
- Collection scheduling
- Shipping label generation
- Tracking integration
- Status updates from logistics partners

### Phase 4: Refund Processing
- Payment gateway integration
- Gift card issuance
- Account credit processing
- Refund confirmation notifications

### Phase 5: Analytics
- Return reason analysis
- Return rate tracking
- Product quality insights
- Fraud detection
