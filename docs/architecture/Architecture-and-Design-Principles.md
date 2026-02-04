# Architecture and Design Principles

## Architectural Style

This project implements **Clean Architecture** (also known as Hexagonal Architecture or Ports and Adapters) with **Domain-Driven Design** (DDD) tactical patterns.

## Core Principles

### Separation of Concerns
Each layer has a single, well-defined responsibility with clear boundaries between layers.

### Dependency Rule
Dependencies point inward towards the domain layer. Outer layers depend on inner layers, never the reverse.

```
interfaces → application → domain
infrastructure → application → domain
```

The domain layer has no dependencies on external frameworks or libraries except for basic utilities (UUID, time).

### Testability
Business logic is isolated in the domain layer, making it easy to test without external dependencies.

## Layer Breakdown

### Domain Layer (`internal/domain/`)

**Purpose**: Core business logic and rules  
**Dependencies**: None (except standard library and basic utilities)

#### Entities (`entities/`)
Core business objects with identity and lifecycle:
- `Conversation`: Customer support interaction session
- `Message`: Individual messages in conversations
- `ReturnRequest`: Return requests with status workflow

Entities contain:
- Business validation logic
- State transition methods
- Invariant enforcement

#### Repository Interfaces (`repositories/`)
Define data persistence contracts without implementation details:
- `ConversationRepository`
- `MessageRepository`
- `ReturnRequestRepository`

Repositories provide:
- CRUD operations
- Query methods
- No knowledge of database technology

#### Service Interfaces (`services/`)
Define external service contracts:
- `OrdersAPIClient`: Interface for Orders API integration

### Application Layer (`internal/application/`)

**Purpose**: Orchestrate use cases and coordinate domain operations  
**Dependencies**: Domain layer only

#### Use Cases (`usecases/`)
Implement specific application operations:
- `CreateConversationUseCase`: Create new conversation with initial messages
- `GetConversationUseCase`: Retrieve conversation with messages
- `SendMessageUseCase`: Add message to existing conversation
- `CreateReturnRequestUseCase`: Submit return request with validation
- `GetReturnRequestUseCase`: Retrieve return request details
- `ListReturnRequestsUseCase`: List return requests for customer

Use cases:
- Orchestrate multiple domain operations
- Coordinate repository interactions
- Enforce business workflows
- Return DTOs, not entities

#### DTOs (`dto/`)
Data Transfer Objects for request/response:
- `ConversationDTO`: Conversation creation and retrieval
- `MessageDTO`: Message payloads
- `ReturnRequestDTO`: Return request data
- `ReturnItemDTO`: Individual return items

DTOs:
- Contain JSON binding tags
- Include validation tags
- Provide clear API contracts
- Separate internal entities from external representation

### Infrastructure Layer (`internal/infrastructure/`)

**Purpose**: Implement external integrations and technical capabilities  
**Dependencies**: Domain interfaces, external libraries

#### Persistence (`persistence/sqlite/`)
SQLite implementations of repository interfaces:
- `ConversationRepository`: SQLite-backed conversation storage
- `MessageRepository`: SQLite-backed message storage
- `ReturnRequestRepository`: SQLite-backed return request storage
- `db.go`: Database connection management

Implementation details:
- JSON marshalling for complex fields (context, metadata, items)
- RFC3339 timestamp handling
- Foreign key enforcement
- Connection pooling (max 25 open, 5 idle)

#### Clients (`clients/`)
External API client implementations:
- `http/orders_api_client.go`: HTTP client for Orders API
- `mock/orders_api_client.go`: Mock client for development/testing

Clients:
- Implement domain service interfaces
- Handle HTTP communication
- Parse external API responses
- Provide error handling

### Interface Layer (`internal/interfaces/http/`)

**Purpose**: Expose application functionality via HTTP  
**Dependencies**: Application layer, Gin framework

#### Handlers (`handlers/`)
HTTP request handlers:
- `ConversationHandler`: Conversation endpoints
- `ReturnHandler`: Return request endpoints

Handlers:
- Parse HTTP requests into DTOs
- Invoke use cases
- Format responses as JSON
- Handle HTTP errors

## Design Patterns

### Repository Pattern
Abstracts data persistence, allowing domain layer to remain persistence-ignorant.

**Example**:
```go
type ConversationRepository interface {
    Create(ctx context.Context, conversation *entities.Conversation) error
    GetByID(ctx context.Context, id string) (*entities.Conversation, error)
    Update(ctx context.Context, conversation *entities.Conversation) error
}
```

### Use Case Pattern
Encapsulates single application operation with clear input/output.

**Example**:
```go
type CreateConversationUseCase struct {
    conversationRepo repositories.ConversationRepository
    messageRepo      repositories.MessageRepository
}
```

### Dependency Injection
Dependencies injected via constructors, enabling testability and flexibility.

**Example** (from `main.go`):
```go
conversationRepo := sqlite.NewConversationRepository(db)
messageRepo := sqlite.NewMessageRepository(db)
createConvUseCase := usecases.NewCreateConversationUseCase(conversationRepo, messageRepo)
```

## Coding Standards

### File Organisation
- One primary type per file
- Test files alongside implementation (`*_test.go`)
- Interfaces in separate files from implementations

### Naming Conventions
- Entities: Singular nouns (`Conversation`, `ReturnRequest`)
- Repositories: `<Entity>Repository`
- Use cases: `<Action><Entity>UseCase`
- Constructors: `New<Type>`
- Getters: No `Get` prefix for simple field access

### Error Handling
- Custom error types in `pkg/errors/`
- Errors include context and error codes
- Use case errors wrap domain errors with application context

**Example**:
```go
type AppError struct {
    Code    string
    Message string
    Err     error
}
```

### Validation
- Entity validation in domain layer
- Input validation in DTOs using struct tags
- Use case validation for cross-entity rules

### Comments
Comments used only for:
- Complex logic that isn't self-evident
- Public API documentation
- Business rule clarification

Avoid:
- Redundant comments explaining obvious code
- Inline comments for self-documenting code

### Testing Requirements
- Unit tests for all domain entities
- Use case tests with mocked repositories
- Integration tests for repository implementations
- Minimum 70% code coverage for new code

## Database Design

### Schema Principles
- Normalised schema (3NF)
- JSON columns for flexible nested data (context, metadata, items)
- Foreign keys with appropriate cascade behaviour
- Indices on frequently queried columns

### Migration Strategy
- SQL migration files in `migrations/`
- Sequential numbering (`001_`, `002_`, etc.)
- Up migrations only in MVP
- Manual migration execution via Makefile

## API Design

### REST Principles
- Resource-based URLs (`/api/v1/conversations`, `/api/v1/returns`)
- HTTP verbs for operations (GET, POST, PUT, DELETE)
- JSON request/response bodies
- Appropriate HTTP status codes

### Versioning
- URL path versioning (`/api/v1/`)
- Major version in path
- Breaking changes require new version

### Error Responses
Consistent error format:
```json
{
  "error": {
    "code": "VALIDATION_FAILED",
    "message": "Descriptive error message"
  }
}
```

## Dependency Management

### External Dependencies
Minimal external dependencies:
- **Gin**: HTTP framework (widely adopted, stable)
- **modernc.org/sqlite**: Pure Go SQLite (no CGo required)
- **UUID**: Standard UUID generation
- **JWT**: Authentication (planned)
- **godotenv**: Environment configuration

### Dependency Rules
- Domain layer: No external dependencies beyond standard library
- Application layer: Domain layer only
- Infrastructure layer: Implements domain interfaces using external libraries
- Interface layer: Application layer plus web framework

## Configuration Management

### Environment Variables
Configuration via `.env` file or environment variables:
- `PORT`: HTTP server port (default: 8080)
- `DB_PATH`: SQLite database path (default: ./data/chatbot.db)
- `LOG_LEVEL`: Logging level (debug/info/error)
- `ORDERS_API_URL`: Orders API endpoint (mock if empty)

### Configuration Loading
- `godotenv` loads `.env` file at startup
- Fallback to defaults if variables not set
- No sensitive data in version control

## Future Architectural Considerations

### Authentication Layer
- JWT middleware in infrastructure layer
- Customer authentication via Bearer token
- User context injection into requests

### Caching Layer
- In-memory cache for frequently accessed data
- Cache interface in domain/services
- Implementation in infrastructure layer

### Message Queue Integration
- Async processing for long-running operations
- Event publishing for domain events
- Queue interface in domain layer
