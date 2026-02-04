# Orders API Integration

## Overview

The Orders API integration provides access to customer order data for validation and refund calculation during the returns process. The integration is defined through a domain service interface with implementations for both mock (development) and HTTP (production) clients.

## Purpose

The Orders API integration serves several critical functions:
1. **Order Validation**: Verify orders exist and belong to the customer
2. **Item Validation**: Confirm items are part of the order and eligible for return
3. **Refund Calculation**: Determine refund amounts based on returned items
4. **Order Data Retrieval**: Fetch order details for display and processing

## Architecture

### Domain Interface

**Location**: `internal/domain/services/orders_api_client.go`

```go
type OrdersAPIClient interface {
    GetOrderByNumber(ctx context.Context, orderNumber string) (*models.Order, error)
    GetOrderByID(ctx context.Context, orderID string) (*models.Order, error)
    ValidateOrderItems(ctx context.Context, orderNumber string, itemIDs []string) error
    CalculateRefundAmount(ctx context.Context, orderNumber string, itemIDs []string) (int, error)
}
```

This interface is defined in the domain layer, keeping the core business logic independent of implementation details.

### Implementations

**Mock Client** (MVP):
- **Location**: `internal/infrastructure/clients/mock/orders_api_client.go`
- **Purpose**: Development and testing
- **Data**: In-memory sample orders
- **Active**: Default in MVP when `ORDERS_API_URL` is not set

**HTTP Client** (Future):
- **Location**: `internal/infrastructure/clients/http/orders_api_client.go`
- **Purpose**: Production integration
- **Transport**: HTTP REST API
- **Active**: When `ORDERS_API_URL` environment variable is set

## API Operations

### GetOrderByNumber

**Purpose**: Retrieve order details by order number

**Parameters**:
- `orderNumber`: Order identifier (e.g., "ORD-123456")

**Returns**: `*models.Order` or error if not found

**Usage**:
```go
order, err := ordersAPIClient.GetOrderByNumber(ctx, "ORD-123456")
```

**Use Cases**:
- Return request creation
- Order ownership verification
- Display order details to customer

### GetOrderByID

**Purpose**: Retrieve order details by internal order ID

**Parameters**:
- `orderID`: Internal order identifier

**Returns**: `*models.Order` or error if not found

**Usage**:
```go
order, err := ordersAPIClient.GetOrderByID(ctx, "order-123")
```

**Use Cases**:
- Alternative order lookup
- Internal system operations

### ValidateOrderItems

**Purpose**: Verify items exist in order and can be returned

**Parameters**:
- `orderNumber`: Order identifier
- `itemIDs`: Array of order item IDs to validate

**Returns**: Error if validation fails, nil if all items valid

**Validation Checks**:
1. Order exists
2. All item IDs exist in the order
3. Items are marked as returnable
4. Order status allows returns (e.g., delivered)

**Usage**:
```go
itemIDs := []string{"ITEM-1", "ITEM-2"}
err := ordersAPIClient.ValidateOrderItems(ctx, "ORD-123456", itemIDs)
```

**Use Cases**:
- Return request validation
- Prevent invalid return submissions

### CalculateRefundAmount

**Purpose**: Calculate total refund amount for specified items

**Parameters**:
- `orderNumber`: Order identifier
- `itemIDs`: Array of order item IDs to refund

**Returns**: Refund amount in cents

**Calculation**:
- Sums item prices for specified items
- Does not include shipping (business rule)
- Returns amount in cents (e.g., 15999 = $159.99)

**Usage**:
```go
amount, err := ordersAPIClient.CalculateRefundAmount(ctx, "ORD-123456", itemIDs)
```

**Use Cases**:
- Display refund amount to customer
- Store refund amount in return request
- Financial reporting

## Order Data Model

### Order Structure

**Location**: `internal/domain/models/order.go`

```go
type Order struct {
    ID              string
    OrderNumber     string
    CustomerID      string
    Status          OrderStatus
    Items           []OrderItem
    SubtotalCents   int
    ShippingCents   int
    TaxCents        int
    TotalCents      int
    CurrencyCode    string
    OrderDate       time.Time
    DeliveryDate    *time.Time
    ShippingAddress Address
}
```

### Order Item Structure

```go
type OrderItem struct {
    ID           string
    ProductID    string
    ProductName  string
    SKU          string
    Quantity     int
    PriceCents   int
    TotalCents   int
    ImageURL     string
    IsReturnable bool
}
```

### Order Status

| Status | Description | Can Return |
|--------|-------------|------------|
| `pending` | Order placed, not confirmed | No |
| `confirmed` | Order confirmed, being processed | No |
| `shipped` | Order shipped to customer | Yes |
| `delivered` | Order delivered to customer | Yes |
| `cancelled` | Order cancelled | No |

**Business Rule**: Only shipped and delivered orders allow returns.

## Mock Client Implementation

### Purpose

Provides realistic order data for development and testing without requiring external API connectivity.

### Sample Orders

The mock client seeds two sample orders on initialisation:

**Order 1**: `ORD-123456`
- Customer: `demo-customer-123`
- Status: `delivered`
- Items:
  - Nike Air Max 270 ($159.99)
  - Adidas Running Socks 3-pack ($19.99)
- Total: $201.18

**Order 2**: `ORD-789012`
- Customer: `demo-customer-123`
- Status: `delivered`
- Items:
  - Puma T-Shirt x2 ($29.99 each)
- Total: $70.83

### Validation Logic

**Item Existence**:
- Checks if item ID exists in order items
- Returns error if item not found

**Returnability**:
- Verifies `IsReturnable` flag is true
- Checks order status is `shipped` or `delivered`
- Returns error if conditions not met

**Refund Calculation**:
- Sums `PriceCents` for each specified item
- Returns total in cents

### Limitations

- In-memory only (data lost on restart)
- Fixed sample data
- No actual API calls
- Customer always `demo-customer-123`

## HTTP Client Implementation

### Configuration

**Environment Variable**: `ORDERS_API_URL`
- **Example**: `https://api.bash.co.za/orders/v1`
- **Default**: Empty (uses mock client)

**Activation**:
```bash
export ORDERS_API_URL="https://api.bash.co.za/orders/v1"
```

### Planned Endpoints

**Get Order by Number**:
```
GET {ORDERS_API_URL}/orders/number/{orderNumber}
Authorization: Bearer {token}
```

**Get Order by ID**:
```
GET {ORDERS_API_URL}/orders/{orderID}
Authorization: Bearer {token}
```

**Validate Order Items** (may be client-side logic):
```
GET {ORDERS_API_URL}/orders/{orderNumber}/items/validate
Body: { "item_ids": ["ITEM-1", "ITEM-2"] }
```

**Calculate Refund** (may be dedicated endpoint):
```
POST {ORDERS_API_URL}/orders/{orderNumber}/calculate-refund
Body: { "item_ids": ["ITEM-1", "ITEM-2"] }
```

### Authentication

Future HTTP implementation will use:
- Bearer token authentication
- Customer-scoped access
- Token passed from incoming request

### Error Handling

**HTTP Status Mapping**:
- `404 Not Found` → `NotFoundError`
- `400 Bad Request` → `InvalidInputError`
- `403 Forbidden` → `ForbiddenRequestError`
- `500 Server Error` → `DatabaseError` (generic error)

### Retry Logic (Future)

- Automatic retry for transient failures
- Exponential backoff
- Circuit breaker pattern

## Integration in Returns Flow

### Return Request Creation Flow

1. **Customer submits return request** with order number and items
2. **Get order data**: `GetOrderByNumber(orderNumber)`
3. **Verify ownership**: Check `order.CustomerID == customerID`
4. **Validate items**: `ValidateOrderItems(orderNumber, itemIDs)`
5. **Calculate refund**: `CalculateRefundAmount(orderNumber, itemIDs)`
6. **Create return request** with validated data
7. **Return response** with refund amount

### Error Scenarios

**Order not found**:
- Error: `NotFoundError`
- User message: "Order not found"
- Action: Reject return request

**Ownership mismatch**:
- Error: `InvalidInputError`
- User message: "Order does not belong to customer"
- Action: Reject return request (security)

**Invalid items**:
- Error: `InvalidInputError`
- User message: "Item {id} not found in order" or "Item cannot be returned"
- Action: Reject return request with details

**API unavailable**:
- Error: `DatabaseError`
- User message: "Service temporarily unavailable"
- Action: Return 503 error, retry later

## Configuration

### Environment Variables

```bash
# Orders API Configuration
ORDERS_API_URL=""  # Empty uses mock client
# ORDERS_API_URL="https://api.bash.co.za/orders/v1"  # HTTP client

# Authentication (future)
ORDERS_API_KEY=""
ORDERS_API_TIMEOUT="30s"
```

### Client Selection Logic

**Location**: `cmd/api/main.go`

```go
ordersAPIURL := os.Getenv("ORDERS_API_URL")
var ordersAPIClient services.OrdersAPIClient

if ordersAPIURL == "" || ordersAPIURL == "mock" {
    log.Println("Using mock Orders API client")
    ordersAPIClient = mock.NewMockOrdersAPIClient()
} else {
    log.Printf("Using HTTP Orders API client: %s", ordersAPIURL)
    ordersAPIClient = http.NewOrdersAPIClient(ordersAPIURL)
}
```

## Testing Strategy

### Unit Tests

**Mock Client Tests**:
- Verify sample data seeding
- Test order retrieval by number and ID
- Validate item validation logic
- Test refund calculation accuracy

**HTTP Client Tests** (Future):
- Mock HTTP responses
- Test error handling
- Verify authentication headers
- Test retry logic

### Integration Tests

**With Mock Client**:
- Full return request creation flow
- Order validation in use cases
- Error scenarios

**With HTTP Client** (Future):
- Real API integration tests
- Authentication flow
- Performance testing
- Failover scenarios

## Future Enhancements

### Phase 2: Production API

- Implement HTTP client
- Add authentication
- Error handling and retry logic
- Monitoring and logging

### Phase 3: Caching

- Cache order data to reduce API calls
- Cache invalidation strategy
- TTL configuration

### Phase 4: Advanced Features

- Batch order retrieval
- Order history pagination
- Real-time order status updates
- Webhook integration for order events

### Phase 5: Performance

- Connection pooling
- Request batching
- Circuit breaker implementation
- Rate limiting

## Related Documentation

- [bff-context.md](../bff-context.md): Original BFF integration specification
- [Returns Feature](../features/Returns-Feature.md): Returns processing that uses this integration
- [Architecture and Design Principles](../architecture/Architecture-and-Design-Principles.md): Dependency injection and interface design
