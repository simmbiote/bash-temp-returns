# Database Schema

## Overview

The database schema supports the customer support chatbot API using SQLite for data persistence. The schema is designed following normalisation principles with appropriate indices for performance and foreign keys for referential integrity.

## Database Technology

**Database**: SQLite 3  
**Implementation**: modernc.org/sqlite v1.44.3 (pure Go, no CGo)  
**Location**: `./data/chatbot.db` (configurable via `DB_PATH` environment variable)  
**Migration**: SQL file at `migrations/001_sqlite_schema.sql`

## Schema Design Principles

1. **Normalisation**: Schema follows 3NF to minimise data redundancy
2. **JSON for Complex Data**: Flexible nested data stored as JSON (context, metadata, items)
3. **Timestamp Consistency**: All timestamps in RFC3339 format
4. **Foreign Keys**: Referential integrity enforced with appropriate cascade behaviour
5. **Indices**: Strategic indices on frequently queried columns
6. **Check Constraints**: Enum-like constraints for status and type fields

## Core Tables (MVP)

### conversations

Stores customer support conversation sessions.

**Columns**:

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| `id` | TEXT | PRIMARY KEY | UUID identifier |
| `customer_id` | TEXT | NOT NULL | Customer identifier |
| `type` | TEXT | NOT NULL, CHECK | `natural_language` or `guided_flow` |
| `intent` | TEXT | NULL | Classified intent (return, refund, etc.) |
| `current_step` | TEXT | NULL | Current step in guided flow |
| `flow_id` | TEXT | NULL | Flow definition identifier |
| `context` | TEXT | NOT NULL, DEFAULT '{}' | JSON key-value context store |
| `created_at` | TEXT | NOT NULL, DEFAULT now | Creation timestamp (RFC3339) |
| `updated_at` | TEXT | NOT NULL, DEFAULT now | Last update timestamp (RFC3339) |
| `completed_at` | TEXT | NULL | Completion timestamp (RFC3339) |

**Constraints**:
- `type` must be `natural_language` or `guided_flow`

**Indices**:
- `idx_conversations_customer_id`: Fast lookup by customer
- `idx_conversations_created_at`: Chronological ordering

**Relationships**:
- One-to-many with `messages` (CASCADE delete)
- One-to-one optional with `return_requests` (SET NULL on delete)

**JSON Fields**:
- `context`: Flexible key-value pairs for conversation state
  ```json
  {
    "order_number": "ORD-12345",
    "return_reason": "wrong_size",
    "selected_items": ["ITEM-1", "ITEM-2"]
  }
  ```

---

### messages

Stores individual messages within conversations.

**Columns**:

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| `id` | TEXT | PRIMARY KEY | UUID identifier |
| `conversation_id` | TEXT | NOT NULL, FK | Parent conversation ID |
| `role` | TEXT | NOT NULL, CHECK | `user`, `assistant`, or `system` |
| `content` | TEXT | NOT NULL | Message content |
| `metadata` | TEXT | NOT NULL, DEFAULT '{}' | JSON metadata |
| `created_at` | TEXT | NOT NULL, DEFAULT now | Creation timestamp (RFC3339) |

**Constraints**:
- `role` must be `user`, `assistant`, or `system`
- Foreign key to `conversations(id)` with CASCADE delete

**Indices**:
- `idx_messages_conversation_id`: Fast retrieval of conversation messages
- `idx_messages_created_at`: Chronological ordering

**Relationships**:
- Many-to-one with `conversations` (deleted when conversation deleted)

**JSON Fields**:
- `metadata`: Additional message information
  ```json
  {
    "platform": "mobile_app",
    "version": "2.1.0",
    "intent_confidence": 0.95,
    "entities": ["order_number"]
  }
  ```

---

### return_requests

Stores customer return requests.

**Columns**:

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| `id` | TEXT | PRIMARY KEY | UUID identifier |
| `conversation_id` | TEXT | NULL, FK | Associated conversation (optional) |
| `customer_id` | TEXT | NOT NULL | Customer identifier |
| `order_number` | TEXT | NOT NULL | Original order number |
| `items` | TEXT | NOT NULL | JSON array of return items |
| `reason` | TEXT | NOT NULL | Return reason |
| `detailed_reason` | TEXT | NULL | Additional reason details |
| `photos` | TEXT | NULL | JSON array of photo URLs |
| `refund_method` | TEXT | NOT NULL | `original_payment`, `gift_card`, or `bash_account` |
| `delivery_method` | TEXT | NOT NULL | `collect` or `ship` |
| `collection_point` | TEXT | NULL | Collection location (if collect) |
| `shipping_address` | TEXT | NULL | Return shipping address (if ship) |
| `status` | TEXT | NOT NULL, DEFAULT 'draft' | Current status |
| `estimated_refund_date` | TEXT | NULL | Expected refund date (RFC3339) |
| `refund_amount_cents` | INTEGER | NOT NULL | Refund amount in cents |
| `return_reference` | TEXT | NULL | Return tracking reference |
| `created_at` | TEXT | NOT NULL, DEFAULT now | Creation timestamp (RFC3339) |
| `updated_at` | TEXT | NOT NULL, DEFAULT now | Last update timestamp (RFC3339) |

**Constraints**:
- Foreign key to `conversations(id)` with SET NULL on delete

**Indices**:
- `idx_return_requests_customer_id`: Customer's return history
- `idx_return_requests_order_number`: Returns by order
- `idx_return_requests_status`: Filter by status
- `idx_return_requests_created_at`: Chronological ordering

**Relationships**:
- Many-to-one optional with `conversations`

**JSON Fields**:

- `items`: Array of return items
  ```json
  [
    {
      "order_item_id": "ITEM-1",
      "product_id": "PROD-123",
      "quantity": 1
    }
  ]
  ```

- `photos`: Array of photo URLs
  ```json
  ["https://example.com/photo1.jpg", "https://example.com/photo2.jpg"]
  ```

**Status Values**:
- `draft`: Initial state
- `submitted`: Customer submitted
- `approved`: Approved for processing
- `rejected`: Request rejected
- `collected`: Items collected
- `received`: Items received at warehouse
- `refunded`: Refund processed

---

## Future Tables (Phase 2+)

### refund_requests

Tracks refund processing separate from return requests.

**Purpose**: Manage refund lifecycle independent of return request  
**Status**: Not implemented in MVP  
**Use Cases**: Refund processing, payment gateway integration, refund tracking

**Key Fields**:
- `return_reference`: Links to return request
- `refund_amount_cents`: Amount to refund
- `refund_method`: Payment method
- `status`: `pending`, `processing`, `completed`, `failed`
- `processed_date`: When refund completed
- `failure_reason`: Error details if failed

**Indices**:
- Customer ID, order number, status, return reference

---

### account_actions

Stores customer account update requests with OTP verification.

**Purpose**: Secure account modifications via chatbot  
**Status**: Not implemented in MVP  
**Use Cases**: Email change, password reset, address updates

**Key Fields**:
- `action_type`: Type of account action
- `otp_code`: One-time password for verification
- `otp_expires_at`: OTP expiry timestamp
- `otp_attempts`: Failed verification attempts
- `action_data`: JSON of action details
- `status`: `pending`, `verified`, `completed`, `failed`

**Security Features**:
- OTP expiry enforcement
- Attempt limiting
- Verification timestamp tracking

**Indices**:
- Customer ID, status, action type, creation date

---

### flows

Stores guided flow definitions for structured conversations.

**Purpose**: Define and version guided conversation flows  
**Status**: Table exists, functionality not implemented in MVP  
**Use Cases**: Returns flow, account update flow, order tracking flow

**Key Fields**:
- `name`: Human-readable flow name
- `intent`: Associated intent
- `version`: Flow version number
- `steps`: JSON definition of flow steps
- `is_active`: Only one active flow per intent

**Versioning**:
- Multiple versions can exist
- Only one version active per intent
- Unique index on `(intent)` WHERE `is_active = 1`

**JSON Structure** (example):
```json
{
  "steps": [
    {
      "id": "select_order",
      "prompt": "Which order would you like to return?",
      "type": "order_selection"
    },
    {
      "id": "select_items",
      "prompt": "Select items to return",
      "type": "item_selection"
    }
  ]
}
```

**Indices**:
- Intent, active status, unique active-intent combination

---

### audit_logs

Audit trail for tracking entity changes.

**Purpose**: Compliance, debugging, security  
**Status**: Table exists, logging not implemented in MVP  
**Use Cases**: Track modifications, investigate issues, compliance reporting

**Key Fields**:
- `entity_type`: Type of entity (conversation, return_request, etc.)
- `entity_id`: ID of modified entity
- `action`: Action performed (create, update, delete)
- `user_id`: Who performed action
- `changes`: JSON of field changes

**JSON Structure** (example):
```json
{
  "before": {"status": "submitted"},
  "after": {"status": "approved"},
  "changed_fields": ["status"]
}
```

**Indices**:
- Composite index on entity type and ID
- Creation date for time-based queries

---

## Relationships Diagram

```
┌─────────────────┐
│  conversations  │
├─────────────────┤
│ id (PK)         │
│ customer_id     │◄──────────┐
│ type            │           │
│ intent          │           │
└────────┬────────┘           │
         │                     │
         │ 1:N                 │
         │                     │
    ┌────▼────────┐            │
    │  messages   │            │
    ├─────────────┤            │
    │ id (PK)     │            │
    │ conv_id (FK)│            │
    │ role        │            │
    └─────────────┘            │
                               │
         │ 0..1:1              │
         │                     │
┌────────▼─────────────┐       │
│  return_requests     │       │
├──────────────────────┤       │
│ id (PK)              │       │
│ conversation_id (FK) │───────┘
│ customer_id          │
│ order_number         │
│ status               │
└──────────────────────┘
```

## Data Types and Storage

### Text Storage

SQLite `TEXT` type used for:
- UUIDs (string representation)
- Customer IDs, order numbers
- Enum-like values (status, type, role)
- Timestamps (RFC3339 format: `2026-02-04T10:30:00Z`)
- JSON data (serialised strings)

### Integer Storage

SQLite `INTEGER` type used for:
- Amounts in cents (avoid float precision issues)
- Counters (OTP attempts, versions)
- Boolean flags (is_active: 0 or 1)

### JSON Storage

Complex nested data stored as JSON text:
- Conversation context
- Message metadata
- Return request items
- Flow step definitions
- Audit log changes

**Advantages**:
- Flexibility for varying structures
- No schema changes for new fields
- Easy serialisation/deserialisation in Go

**Trade-offs**:
- Cannot query JSON contents with indices (SQLite JSON1 extension not used)
- Must deserialise for processing

## Indices Strategy

### Primary Indices

Every table has a primary key on `id` (UUID).

### Foreign Key Indices

Indices automatically created on foreign key columns:
- `conversation_id` in messages and return_requests
- Fast JOIN operations
- Efficient cascade deletes

### Query Optimisation Indices

**Customer Lookups**:
- `customer_id` indices on conversations, return_requests, refund_requests, account_actions
- Support "get all for customer" queries

**Status Filtering**:
- `status` indices on return_requests, refund_requests, account_actions
- Enable efficient filtering by processing state

**Time-Based Queries**:
- `created_at` indices on all tables with timestamps
- Support chronological ordering and date range queries

**Unique Constraints**:
- `idx_flows_active_intent`: Ensure only one active flow per intent

## Foreign Key Cascade Behaviour

### CASCADE Delete

**messages → conversations**:
- When conversation deleted, all messages automatically deleted
- Maintains referential integrity
- Prevents orphaned messages

### SET NULL

**return_requests → conversations**:
- When conversation deleted, `conversation_id` set to NULL in return requests
- Preserves return request data
- Conversation link optional

**refund_requests → conversations**:
- Similar behaviour
- Refund data preserved

**account_actions → conversations**:
- Action record preserved
- Conversation link optional

## Database Connection Management

**Configuration**:
- **Max Open Connections**: 25
- **Max Idle Connections**: 5
- **Connection Lifetime**: Unlimited
- **Foreign Keys**: Enabled

**Location**: `internal/infrastructure/persistence/sqlite/db.go`

```go
db.SetMaxOpenConns(25)
db.SetMaxIdleConns(5)
db.Exec("PRAGMA foreign_keys = ON")
```

## Migration Strategy

### Current Approach (MVP)

**File**: `migrations/001_sqlite_schema.sql`  
**Execution**: Manual via Makefile

```bash
make db-create  # Creates database with schema
make db-reset   # Drops and recreates database
```

**Workflow**:
1. SQL file contains all CREATE TABLE statements
2. Executed once to initialise database
3. No versioning or rollback in MVP

### Future Migration Strategy

**Phase 2 Enhancements**:
- Migration versioning system
- Up/down migrations
- Schema version tracking
- Automated migration on startup
- Migration rollback capability

**Potential Tools**:
- golang-migrate/migrate
- pressly/goose
- Custom migration runner

## Backup and Recovery

**Development**:
- SQLite file at `./data/chatbot.db`
- Simple file copy for backup
- Version control excludes database file (`.gitignore`)

**Production** (Future):
- Regular automated backups
- Point-in-time recovery
- Backup retention policy
- Disaster recovery procedures

## Performance Considerations

### Query Performance

**Indexed Columns**: All frequently queried columns have indices  
**JSON Queries**: Not indexed, requires full deserialisation  
**Join Performance**: Foreign key indices support efficient joins

### Connection Pool

**Pool Size**: 25 concurrent connections sufficient for expected load  
**Idle Connections**: 5 kept ready for new requests

### Scalability Limits

SQLite appropriate for:
- Single-instance deployment
- Low to moderate traffic
- Read-heavy workloads
- Embedded applications

Consider migration to PostgreSQL/MySQL for:
- Multi-instance deployment
- High concurrent writes
- Large datasets (> 1M records)
- Advanced query features

## Data Retention

**Current**: No automatic deletion (MVP)

**Future Policies**:
- Archive completed conversations after 90 days
- Delete draft returns after 30 days of inactivity
- Retain audit logs for 2 years
- GDPR compliance: customer data deletion on request

## Schema Validation

**Entity Validation**: Domain entities validate data before persistence  
**Database Constraints**: CHECK constraints enforce enum-like values  
**Application Layer**: Additional validation in use cases

**Validation Layers**:
1. DTO input validation (Gin binding tags)
2. Domain entity validation
3. Database constraints
4. Repository error handling

## Future Schema Enhancements

### Phase 2: Refund Processing
- Implement refund_requests table fully
- Add payment gateway integration fields
- Track refund transaction IDs

### Phase 3: Account Management
- Implement account_actions table
- OTP verification workflow
- Security audit logging

### Phase 4: Guided Flows
- Implement flows table functionality
- Flow execution tracking
- Flow analytics

### Phase 5: Advanced Features
- Conversation attachments table
- Customer feedback table
- Knowledge base integration
- Multi-language support fields
