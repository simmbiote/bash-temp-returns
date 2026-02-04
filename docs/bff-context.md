# Customer Support AI - Orders Integration Technical Specification

## Overview

This document specifies the technical requirements for creating a new Customer Support AI Agent API that integrates with the existing Orders domain to query customer orders and log returns.

## System Architecture

### Current Mobile BFF Orders Implementation

The mobile-bff currently implements order management through:

- **Controller**: [order.controller.ts](apps/mobile-bff/src/order/controllers/order.controller.ts)
- **Domain Library**: `libs/order` (provides `OnlineOrderServiceInterface`)
- **Architecture**: 3-layer DDD pattern (Controller → Service → API/Infrastructure)

### Integration Points

**Primary Service**: `OnlineOrderServiceInterface` from `ecom/order`

**Key APIs**:
- `OrderEntityHttpApi` - Backoffice HTTP API for order retrieval
- `VtexOrderApi` - VTEX order system integration
- `ParcelTrackingApi` - Shipment tracking data

## Authentication & Authorization

### Required Guards

```typescript
@UseGuards(CustomerAuthGuard)  // From 'core' library
```

### User Context

```typescript
interface AuthenticatedUser {
  email: string;  // Used as shopperId
  // Additional identity fields...
}
```

### Authorization Flow

1. **Bearer Token**: All requests require `Authorization: Bearer <token>` header
2. **User Extraction**: AuthGuard injects `AuthenticatedUser` into request object
3. **Shopper Validation**: `shopperId` derived from user email, validated against order ownership
4. **Forbidden Check**: `ForbiddenRequestException` thrown if user doesn't own the order

**Critical**: Order detail requests validate that `user.email` matches order's `customer.idKey`

## Core Endpoints to Replicate

### 1. Get Order History

**Endpoint**: `GET /order/orders/history`

**Controller Implementation**:
```typescript
@Get('/orders/history')
@ApiResponse({ type: OrderList })
async orderHistory(
  @Request() req: any,
  @Query() query: OrderHistoryDto,
): Promise<OrderList>
```

**Request DTO**: `OrderHistoryDto`
```typescript
{
  page?: number;       // Default: 0
  pageSize?: number;   // Default: 20
}
```

**Response Model**: `OrderList`
```typescript
{
  pages: number;
  total: number;
  page: number;
  pageSize: number;
  orders: OrderSummary[];  // Array of order summaries
}
```

**Service Method**: 
```typescript
OnlineOrderServiceInterface.searchOrders(
  params: SearchOrdersDto,
  authToken: string
): Promise<OrderList>
```

**Key Fields in OrderSummary**:
- `id` - Order group number
- `itemsTotal` - Number of items
- `subtotalAmountCents` - Order subtotal (cents)
- `discountAmountCents` - Discount applied (cents)
- `shippingAmountCents` - Shipping cost (cents)
- `items: OrderItem[]` - Individual order items
- `deliveryInfo` - Delivery/pickup location info

### 2. Get Detailed Order

**Endpoint**: `GET /order/:orderNumber`

**Controller Implementation**:
```typescript
@Get('/:orderNumber')
@ApiResponse({ type: OnlineDetailedOrder })
async getOrder(
  @Request() req: Request,
  @Param() params: GetDetailedOrderParamDto,
): Promise<OnlineDetailedOrder>
```

**Request DTO**: `GetDetailedOrderParamDto`
```typescript
{
  orderNumber: string;  // Order identifier
}
```

**Response Model**: `OnlineDetailedOrder`
```typescript
{
  // Order totals
  subtotal: number;
  discount: number;
  shippingCost: number;
  
  // Fulfillment info
  fulfilledIn: string;
  deliveryDescription?: string | null;
  
  // Parcel & shipping
  parcels: DetailedOrderParcel[];
  invoices: DetailedOrderInvoice[];
  shippingAddress: DetailedOrderShipping;
  shippingOption: DetailedOrderShippingOption;
  
  // Customer info
  customerDetails: CustomerDetails;
  
  // Return eligibility
  isEligibleForReturn: boolean;
}
```

**Service Method**:
```typescript
OnlineOrderServiceInterface.getDetailedOrder(
  data: GetDetailedOrderDto
): Promise<OnlineDetailedOrder>
```

**Key Fields for Customer Support**:
- `isEligibleForReturn` - Whether order can be returned
- `parcels[]` - Tracking info for shipments
- `customerDetails` - Customer contact information
- Return-related fields in items:
  - `returnCode`
  - `returnReason`
  - `returnReference`
  - `returnItemStatus`
  - `refundProcessed`
  - `allowRefund`

### 3. Get Parcel Tracking

**Endpoint**: `GET /order/parcel/:parcelId`

**Controller Implementation**:
```typescript
@Get('/parcel/:parcelId')
@ApiResponse({ type: Object })
async getParcelTracking(
  @Request() req: any,
  @Param() params: GetParcelTrackingParamDto,
): Promise<any>
```

**Request DTO**: `GetParcelTrackingParamDto`
```typescript
{
  parcelId: string;
}
```

**Service Method**:
```typescript
OnlineOrderServiceInterface.getParcelTracking(
  parcelId: string,
  authToken: string
): Promise<any>
```

## Returns Logging Requirements

### Current State

**⚠️ No Write Operations Exist**: The current orders implementation is **READ-ONLY**. There are no endpoints for:
- Creating returns
- Updating order status
- Logging refunds

### Required New Functionality

To enable returns logging, the following must be implemented:

#### 1. New Service Method

```typescript
interface OnlineOrderServiceInterface {
  // Add to existing interface
  logReturn(
    dto: LogReturnDto,
    authToken: string
  ): Promise<ReturnLoggedModel>;
}
```

#### 2. New DTOs

**LogReturnDto** (Request):
```typescript
export class LogReturnDto {
  @ApiProperty()
  orderNumber: string;
  
  @ApiProperty()
  itemIds: string[];  // Items being returned
  
  @ApiProperty()
  returnReason: string;
  
  @ApiPropertyOptional()
  returnNotes?: string;
  
  @ApiProperty()
  refundMethod: 'ORIGINAL_PAYMENT' | 'STORE_CREDIT' | 'TFG_MONEY';
}
```

**ReturnLoggedModel** (Response):
```typescript
export class ReturnLoggedModel {
  @ApiProperty()
  returnReference: string;
  
  @ApiProperty()
  status: 'LOGGED' | 'PENDING_APPROVAL' | 'APPROVED';
  
  @ApiProperty()
  estimatedRefundDate: string;
  
  @ApiProperty()
  refundAmountCents: number;
}
```

#### 3. Infrastructure API Layer

**New API Required**: Return management API client (e.g., `ReturnManagementApi`)

```typescript
@Injectable()
export class ReturnManagementApi {
  async createReturn(
    request: ReturnManagementRequest,
    authToken: string
  ): Promise<ReturnManagementResponse> {
    // POST to return management system
    // Map to external provider's API contract
  }
}
```

**Note**: Infrastructure layer requires:
- HTTP client configuration (app-level)
- Provider-specific exception handling (`ReturnManagementException`)
- Request/Response models for external API

## Module Dependencies

### Required Libraries

```typescript
import { OnlineOrderServiceInterface, OrderList } from 'ecom/order';
import { OrderHistoryDto } from 'ecom/order/domain/dtos/order-history.dto';
import { OnlineDetailedOrder } from 'ecom/order/domain/models/detailed-order/detailed-order-v1.model';
import { AuthenticatedUser, CustomerAuthGuard } from 'core';
```

### Module Setup

The new Customer Support API app must:

1. **Import OnlineOrderModule**:
```typescript
import { OnlineOrderModule } from 'ecom/order';

@Module({
  imports: [OnlineOrderModule],
  controllers: [CustomerSupportOrderController],
})
export class CustomerSupportModule {}
```

2. **Configure HTTP Clients** (app-level):
   - Backoffice HTTP client (`TokenConstants.backofficeHttpClient`)
   - VTEX HTTP client
   - Return Management HTTP client (new)

3. **Inject Service Interface**:
```typescript
constructor(
  private readonly orderService: OnlineOrderServiceInterface
) {}
```

## Data Flow

### Query Orders Flow

```
Client Request
  ↓
CustomerAuthGuard (validates JWT, extracts user)
  ↓
CustomerSupportOrderController
  ↓
OnlineOrderService (domain logic)
  ↓
OrderEntityHttpApi (infrastructure)
  ↓
Backoffice HTTP API
  ↓
Response Mapper (API → Domain model)
  ↓
Return OrderList to client
```

### Log Return Flow (New)

```
Client Request (LogReturnDto)
  ↓
CustomerAuthGuard (validates JWT, extracts user)
  ↓
CustomerSupportOrderController
  ↓
OnlineOrderService.logReturn()
  ↓
Validate order ownership & eligibility
  ↓
ReturnManagementApi (infrastructure - NEW)
  ↓
External Return Management System
  ↓
Response Mapper (API → Domain model)
  ↓
Return ReturnLoggedModel to client
```

## Error Handling

### Exception Types

```typescript
// From 'core' library
throw new ForbiddenRequestException({
  message: 'User does not own this order',
  meta: { orderNumber, userId }
});
```

### Provider-Specific Exceptions (New)

For return management integration:

```typescript
// libs/order/src/infrastructure/exceptions/return-management.exception.ts
export class ReturnManagementException extends Error {
  constructor(message: string, public statusCode: number) {
    super(message);
  }
}
```

## Security Considerations

### 1. Order Ownership Validation

**Critical**: Always validate `user.email` matches order's `customer.idKey`

```typescript
const idk = this.getIdKeyFromShopperId(data.shopperId);
if (idk !== response?.customer?.idKey) {
  throw new ForbiddenRequestException({
    message: 'Fetch order from backoffice forbidden',
    meta: { data }
  });
}
```

### 2. Token Propagation

Auth token must be passed to infrastructure APIs:

```typescript
const authHeaders = extractAuthTokenFromHeader(req);
await this.orderService.searchOrders(params, authHeaders);
```

### 3. Return Eligibility Check

Before logging returns, validate:
- `order.isEligibleForReturn === true`
- Return window hasn't expired
- Items not already returned

## Testing Requirements

### Controller Tests

```typescript
// Mock CustomerAuthGuard
.overrideGuard(CustomerAuthGuard)
.useValue({ canActivate: () => true })

// Mock user context
req['user'] = { email: 'test@example.com' };
```

### Service Tests

```typescript
// Mock infrastructure APIs
const mockOrderEntityApi = createMock<OrderEntityHttpApi>();
const mockReturnManagementApi = createMock<ReturnManagementApi>();

// Use factories with fixed values (NO Math.random())
const order = BoOrderDetailsFactory.build({
  orderNumber: 'ORD-12345',  // Fixed value
  customer: {
    idKey: 'test-customer',  // Fixed value
  }
});
```

**⚠️ Critical**: Use deterministic test data (fixed values, no `faker.helpers.arrayElement()` or `faker.number.int()` in iteration counts)

## Environment Configuration

### Required Variables

```bash
# Backoffice API
BACKOFFICE_API_URL=https://...
BACKOFFICE_API_KEY=...

# Return Management API (NEW)
RETURN_MANAGEMENT_API_URL=https://...
RETURN_MANAGEMENT_API_KEY=...

# Authentication
JWT_SECRET=...
JWT_EXPIRATION=...
```

**Note**: Use AWS Secrets Manager for sensitive values, Helm `extraEnv` for non-sensitive config

## File Structure

### New App: `apps/customer-support-ai`

```
apps/customer-support-ai/
├── src/
│   ├── order/
│   │   ├── controllers/
│   │   │   └── customer-support-order.controller.ts
│   │   ├── dtos/
│   │   │   ├── log-return.dto.ts
│   │   │   └── index.ts
│   │   └── index.ts
│   ├── network/
│   │   └── network-client.module.ts
│   └── main.ts
├── tests/
│   └── order/
│       └── controllers/
│           └── customer-support-order.controller.spec.ts
├── package.json
├── tsconfig.json
└── nodemon.json
```

### New Library Extensions: `libs/order`

```
libs/order/src/
├── domain/
│   ├── dtos/
│   │   └── log-return.dto.ts (NEW)
│   ├── models/
│   │   └── return-logged.model.ts (NEW)
│   └── services/
│       └── online-order.service.ts (UPDATE - add logReturn method)
└── infrastructure/
    ├── apis/
    │   └── return-management.api.ts (NEW)
    ├── models/
    │   ├── return-management.request.ts (NEW)
    │   └── return-management.response.ts (NEW)
    └── exceptions/
        └── return-management.exception.ts (NEW)
```

## Swagger Documentation

All endpoints must include:

```typescript
@ApiTags('customer-support-orders')
@ApiBearerAuth()
@Controller('/customer-support/orders')

@ApiResponse({ type: OrderList })
@ApiOperation({ summary: 'Get customer order history' })
```

## Naming Conventions

**Strict adherence required** (enforced by `scripts/validate-file-naming.script.ts`):

- Services: `*.service.ts`
- DTOs: `*.dto.ts`
- Models: `*.model.ts`
- Controllers: `*.controller.ts`
- APIs: `*.api.ts`
- Mappers: `*.mapper.ts`
- Interfaces: `*.interface.ts`

## Validation Checklist

Before deployment:

- [ ] Run `make validate-all` - All validations pass
- [ ] Run `make validate-file-naming` - File naming compliant
- [ ] Run `make validate-internal-imports` - Import paths use `ecom/*` aliases
- [ ] Write tests (coverage threshold enforced by Codecov)
- [ ] Document new endpoints in Swagger
- [ ] Configure app-level HTTP clients
- [ ] Add AWS Secrets Manager entries
- [ ] Update Helm charts with environment variables
- [ ] Add Codecov component entry for new app

## References

- Architecture: [docs/architecture.md](docs/architecture.md)
- Code Standards: [docs/code_standards_and_naming.md](docs/code_standards_and_naming.md)
- File Structure: [docs/file_structure_and_naming.md](docs/file_structure_and_naming.md)
- Testing: [docs/testing.md](docs/testing.md)
- Environment Setup: [docs/env_setup.md](docs/env_setup.md)

## Summary

**Read-Only Integration**: Can be implemented immediately by replicating mobile-bff order endpoints with CustomerAuthGuard.

**Returns Logging**: Requires new infrastructure (ReturnManagementApi), service methods, and DTOs to be implemented first.

**Key Principle**: Follow 3-layer architecture strictly - Controller (App) → Service (Domain) → API (Infrastructure).