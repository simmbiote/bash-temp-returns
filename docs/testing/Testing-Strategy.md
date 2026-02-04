# Testing Strategy

## Overview

The testing strategy ensures code quality, correctness, and maintainability through comprehensive unit, integration, and end-to-end testing. The project follows testing best practices aligned with Go conventions and clean architecture principles.

## Testing Philosophy

1. **Test-First Mindset**: Tests guide design and ensure requirements are met
2. **Isolation**: Tests are independent and can run in any order
3. **Fast Feedback**: Unit tests run quickly for rapid development cycles
4. **Realistic Scenarios**: Integration tests use realistic data and workflows
5. **Coverage Goals**: Minimum 70% code coverage for all new code

## Testing Pyramid

```
        /\
       /  \
      / E2E\           Few, slow, high-value
     /------\
    /  Inte- \         Moderate number, medium speed
   /  gration \
  /------------\
 /    Unit      \      Many, fast, focused
/________________\
```

## Test Types

### Unit Tests

**Purpose**: Test individual functions, methods, and entities in isolation  
**Location**: `*_test.go` files alongside implementation  
**Framework**: Go standard `testing` package  
**Coverage Target**: 70%+

**Current Coverage**:
- Domain entities: 70.3%
- 16 tests, all passing

**What to Test**:
- Entity creation and validation
- Business logic methods
- State transitions
- Edge cases and error conditions
- Validation rules

**What NOT to Test**:
- External dependencies (mocked)
- Framework code
- Simple getters/setters

**Example Structure**:
```go
func TestNewConversation(t *testing.T) {
    // Arrange
    customerID := "cust-123"
    
    // Act
    conv := NewConversation(customerID, NaturalLanguage)
    
    // Assert
    if conv.ID == "" {
        t.Error("Expected ID to be generated")
    }
    if conv.CustomerID != customerID {
        t.Errorf("Expected CustomerID to be %s, got %s", customerID, conv.CustomerID)
    }
}
```

---

### Integration Tests

**Purpose**: Test interactions between layers and external dependencies  
**Location**: `tests/integration/` (planned)  
**Framework**: Go standard testing + test database  
**Status**: Not yet implemented

**What to Test**:
- Repository implementations with real database
- Use case orchestration with real repositories
- Orders API client with mock HTTP server
- Database transactions and rollbacks
- Error handling across layers

**Test Database**:
- Separate SQLite database for tests
- Created/destroyed for each test run
- Seeded with test fixtures

**Example Structure**:
```go
func TestCreateReturnRequest_Integration(t *testing.T) {
    // Setup test database
    db := setupTestDB(t)
    defer cleanupTestDB(t, db)
    
    // Seed test data
    seedOrders(t, db)
    
    // Create repositories
    returnRepo := sqlite.NewReturnRequestRepository(db)
    
    // Execute test
    // ...
}
```

---

### End-to-End Tests

**Purpose**: Test complete workflows through HTTP API  
**Location**: `tests/e2e/` (planned)  
**Framework**: Go standard testing + HTTP client  
**Status**: Not yet implemented

**What to Test**:
- Complete return request workflow
- Conversation creation and message flow
- Error responses and status codes
- API contracts and response formats

**Test Approach**:
- Start test server
- Make HTTP requests
- Verify responses
- Check database state

**Example Structure**:
```go
func TestReturnWorkflow_E2E(t *testing.T) {
    // Start test server
    server := startTestServer(t)
    defer server.Close()
    
    // Create conversation
    convResp := createConversation(t, server.URL)
    
    // Submit return request
    returnResp := createReturn(t, server.URL, "ORD-123456")
    
    // Verify return was created
    if returnResp.Status != "approved" {
        t.Errorf("Expected status approved, got %s", returnResp.Status)
    }
}
```

---

## Current Test Coverage

### Domain Layer

**Entities** (`internal/domain/entities/`):

**conversation_test.go**:
- `TestNewConversation`: Verify conversation creation
- `TestConversationValidate`: Validation rules for all scenarios
- `TestSetIntent`: Intent setting behaviour
- `TestUpdateContext`: Context management
- `TestComplete`: Completion logic

**return_request_test.go**:
- `TestNewReturnRequest`: Return request creation
- `TestReturnRequestAddItem`: Item addition
- `TestReturnRequestAddItemInvalidQuantity`: Quantity validation
- `TestReturnRequestSetRefundAmount`: Refund amount setting
- `TestReturnRequestValidate`: Comprehensive validation scenarios
- `TestReturnRequestSubmit`: Submission workflow
- `TestReturnRequestApprove`: Approval process
- `TestReturnRequestStatusTransitions`: All status transitions

**Coverage**: 70.3% (16 tests, all passing)

### Application Layer

**Status**: No tests yet (planned)

**Planned Tests**:
- Use case execution with mocked repositories
- Error handling scenarios
- Business rule enforcement
- DTO transformations

### Infrastructure Layer

**Status**: No tests yet (planned)

**Planned Tests**:
- Repository CRUD operations
- JSON serialisation/deserialisation
- Database connection handling
- Mock vs HTTP Orders API client

### Interface Layer

**Status**: No tests yet (planned)

**Planned Tests**:
- Handler request parsing
- Response formatting
- HTTP status codes
- Error response structure

---

## Test Fixtures

**Purpose**: Reusable test data for consistent testing  
**Location**: `tests/fixtures/` (planned)  
**Format**: JSON or Go structs

**Planned Fixtures**:
- Sample orders
- Sample conversations
- Sample return requests
- Customer profiles

**Example**:
```json
{
  "orders": [
    {
      "order_number": "ORD-TEST-001",
      "customer_id": "test-customer-1",
      "status": "delivered",
      "items": [...]
    }
  ]
}
```

---

## Testing Utilities

### Test Helpers

**Database Setup**:
```go
func setupTestDB(t *testing.T) *sql.DB {
    db, err := sqlite.NewDB(":memory:")
    if err != nil {
        t.Fatalf("Failed to create test database: %v", err)
    }
    return db
}

func cleanupTestDB(t *testing.T, db *sql.DB) {
    if err := db.Close(); err != nil {
        t.Errorf("Failed to close database: %v", err)
    }
}
```

**Data Seeding**:
```go
func seedTestConversation(t *testing.T, repo repositories.ConversationRepository) *entities.Conversation {
    conv := entities.NewConversation("test-customer", entities.NaturalLanguage)
    if err := repo.Create(context.Background(), conv); err != nil {
        t.Fatalf("Failed to seed conversation: %v", err)
    }
    return conv
}
```

**Assertion Helpers**:
```go
func assertNoError(t *testing.T, err error) {
    t.Helper()
    if err != nil {
        t.Fatalf("Expected no error, got: %v", err)
    }
}

func assertEqual(t *testing.T, expected, actual interface{}) {
    t.Helper()
    if expected != actual {
        t.Errorf("Expected %v, got %v", expected, actual)
    }
}
```

---

## Mocking Strategy

### Repository Mocks

**Approach**: Interface-based mocking for repositories  
**Tool**: Manual mocks or `go generate` with mockgen (future)

**Example**:
```go
type mockConversationRepository struct {
    conversations map[string]*entities.Conversation
}

func (m *mockConversationRepository) Create(ctx context.Context, conv *entities.Conversation) error {
    m.conversations[conv.ID] = conv
    return nil
}
```

### External Service Mocks

**Orders API**: Mock client already implemented  
**Location**: `internal/infrastructure/clients/mock/orders_api_client.go`

**Benefits**:
- No external dependencies
- Predictable test data
- Fast test execution
- Controlled error scenarios

---

## Test Execution

### Running Tests

**All Tests**:
```bash
make test
```

**Unit Tests Only**:
```bash
make test-unit
# or
go test ./internal/domain/entities/...
```

**With Coverage**:
```bash
make test-coverage
```

**Specific Package**:
```bash
go test ./internal/domain/entities -v
```

**Specific Test**:
```bash
go test ./internal/domain/entities -run TestNewConversation -v
```

### Makefile Targets

**Current**:
```makefile
test:
    go test ./... -v

test-coverage:
    go test ./... -coverprofile=coverage.out
    go tool cover -html=coverage.out
```

**Planned**:
```makefile
test-unit:
    go test ./internal/domain/... -v

test-integration:
    go test ./tests/integration/... -v

test-e2e:
    go test ./tests/e2e/... -v

test-watch:
    # Watch for changes and run tests
```

---

## Continuous Integration

**Status**: Not yet configured  
**Platform**: GitHub Actions (planned)

**Planned CI Workflow**:
1. Install Go dependencies
2. Run linter (`golangci-lint`)
3. Run unit tests
4. Run integration tests
5. Generate coverage report
6. Fail if coverage < 70%

**Example Workflow** (.github/workflows/test.yml):
```yaml
name: Tests
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - uses: actions/setup-go@v2
        with:
          go-version: 1.24
      - run: make test-coverage
      - run: make lint
```

---

## Coverage Requirements

### Minimum Thresholds

- **Overall**: 70% code coverage
- **Domain Layer**: 80% (business logic is critical)
- **Application Layer**: 70%
- **Infrastructure Layer**: 60% (some boilerplate acceptable)
- **Interface Layer**: 70%

### Coverage Reporting

**Tools**:
- `go test -coverprofile`: Generate coverage data
- `go tool cover -html`: View coverage in browser
- `go tool cover -func`: Coverage by function

**Example Output**:
```
internal/domain/entities/conversation.go:15:    NewConversation        100.0%
internal/domain/entities/conversation.go:29:    Validate               90.0%
internal/domain/entities/conversation.go:45:    SetIntent              100.0%
internal/domain/entities/conversation.go:50:    UpdateContext          100.0%
total:                                          (statements)            70.3%
```

---

## Test Documentation

### Test Naming Convention

**Format**: `Test<FunctionName>_<Scenario>`

**Examples**:
- `TestNewConversation`
- `TestConversationValidate_MissingID`
- `TestReturnRequestSubmit_Success`
- `TestReturnRequestSubmit_ValidationError`

### Table-Driven Tests

**Approach**: Use table-driven tests for multiple scenarios

**Example**:
```go
func TestConversationValidate(t *testing.T) {
    tests := []struct {
        name    string
        conv    *Conversation
        wantErr bool
    }{
        {
            name: "valid natural language conversation",
            conv: &Conversation{...},
            wantErr: false,
        },
        {
            name: "missing ID",
            conv: &Conversation{...},
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := tt.conv.Validate()
            if (err != nil) != tt.wantErr {
                t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

---

## Testing Best Practices

### Do's

✅ Test behaviour, not implementation  
✅ Use descriptive test names  
✅ One assertion concept per test  
✅ Arrange-Act-Assert pattern  
✅ Mock external dependencies  
✅ Test edge cases and error paths  
✅ Keep tests independent  
✅ Use table-driven tests for similar scenarios

### Don'ts

❌ Test private functions directly  
❌ Share state between tests  
❌ Make network calls in unit tests  
❌ Test framework code  
❌ Ignore test failures  
❌ Write tests after implementation (test-first preferred)  
❌ Hard-code test data inline (use fixtures)

---

## Future Testing Enhancements

### Phase 2: Integration Tests

- Repository integration tests with SQLite
- Use case integration tests
- Orders API client integration tests
- Database transaction testing

### Phase 3: E2E Tests

- Complete workflow testing
- HTTP API contract testing
- Error scenario testing
- Performance testing

### Phase 4: Advanced Testing

- Property-based testing (rapid)
- Mutation testing
- Fuzz testing for input validation
- Load testing with k6

### Phase 5: Test Automation

- Pre-commit hooks for tests
- Coverage tracking over time
- Automated test report generation
- Test flakiness detection

---

## Performance Testing

**Status**: Not implemented  
**Tool**: Go benchmarks (`testing.B`)

**Example Benchmark**:
```go
func BenchmarkConversationCreate(b *testing.B) {
    for i := 0; i < b.N; i++ {
        NewConversation("cust-123", NaturalLanguage)
    }
}
```

**Run Benchmarks**:
```bash
go test -bench=. -benchmem ./...
```

---

## Test Metrics

**Current Metrics**:
- Total tests: 16
- Passing: 16 (100%)
- Coverage: 70.3%
- Test execution time: < 1 second

**Target Metrics**:
- Total tests: 100+
- Passing: 100%
- Coverage: 75%+
- Unit test execution: < 5 seconds
- Integration test execution: < 30 seconds
- E2E test execution: < 2 minutes

---

## Related Documentation

- [Architecture and Design Principles](../architecture/Architecture-and-Design-Principles.md): Testability through dependency injection
- [Domain Model](../architecture/Domain-Model.md): Entity validation rules tested
- [Returns Feature](../features/Returns-Feature.md): Return workflow testing requirements
