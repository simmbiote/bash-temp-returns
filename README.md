# Customer Support Chatbot API

A Go-based REST API that powers a comprehensive customer support chatbot, enabling customers to perform self-service tasks through both natural language and structured guided flows.

## Features

- **Returns Processing**: Complete returns flow with order selection, item selection, and refund processing
- **Natural Language Understanding**: AI-powered intent classification
- **Guided Flows**: Structured step-by-step customer interactions
- **Order Integration**: Direct integration with Bash Orders API
- **SQLite Database**: Lightweight, zero-config database for local development
- **JWT Authentication**: Secure customer authentication
- **Comprehensive Testing**: Unit, integration, and E2E tests

## Quick Start

### Prerequisites

- Go 1.21 or higher
- SQLite3 (comes with macOS)
- Make (optional, for convenience commands)

### Installation

```bash
# Clone the repository
cd /Users/simmbiote/Projects/bash-returns

# Install dependencies
make deps

# Create and initialize the database
make db-create

# Run the application
make run
```

The server will start on `http://localhost:8080`

### Test the API

```bash
# Health check
curl http://localhost:8080/health

# Expected response:
# {"status":"ok","database":"sqlite","cache":"memory","version":"0.1.0"}
```

## Project Structure

```
├── cmd/api/              # Application entry point
├── internal/
│   ├── domain/          # Domain entities, repositories, services
│   ├── application/     # Use cases and DTOs
│   ├── infrastructure/  # External services, database, middleware
│   └── interfaces/      # HTTP handlers, routes
├── migrations/          # Database migrations
├── tests/              # Test suites
│   ├── unit/           # Unit tests
│   ├── integration/    # Integration tests
│   ├── e2e/            # End-to-end tests
│   └── fixtures/       # Test fixtures
├── data/               # SQLite database (gitignored)
└── uploads/            # File uploads (gitignored)
```

## Development

### Available Make Commands

```bash
make help              # Show all available commands
make build             # Build the application
make run               # Run the application
make test              # Run all tests
make test-coverage     # Run tests with coverage report
make db-create         # Create and initialize database
make db-reset          # Reset database (WARNING: destroys data)
make clean             # Clean build artifacts
make fmt               # Format code
```

### Running Tests

```bash
# All tests
make test

# Unit tests only
make test-unit

# Integration tests
make test-integration

# E2E tests
make test-e2e

# With coverage
make test-coverage
```

### Database Management

```bash
# Create database
make db-create

# Reset database (destroys all data)
make db-reset

# Inspect database
sqlite3 data/chatbot.db "SELECT name FROM sqlite_master WHERE type='table';"
```

## API Endpoints

### Core Endpoints

- `GET /health` - Health check
- `GET /api/v1/ping` - API connectivity test

### Conversation Management (TODO)

- `POST /api/v1/conversations` - Start a new conversation
- `GET /api/v1/conversations/:id` - Get conversation details
- `POST /api/v1/conversations/:id/messages` - Send a message

### Returns (TODO)

- `POST /api/v1/returns` - Create a return request
- `GET /api/v1/returns/:id` - Get return details

### Orders (TODO)

- `GET /api/v1/orders` - Get order history
- `GET /api/v1/orders/:orderNumber` - Get order details

## Configuration

Configuration is managed through environment variables. Copy `.env.example` to `.env` and update values:

```env
PORT=8080
DB_TYPE=sqlite
DB_PATH=./data/chatbot.db
CACHE_TYPE=memory
ORDERS_API_URL=https://web-api.bash.com
JWT_SECRET=your-secret-here
AI_PROVIDER=keyword
LOG_LEVEL=debug
```

## Architecture

This application follows **Clean Architecture** principles:

- **Domain Layer**: Business entities and logic
- **Application Layer**: Use cases and application logic
- **Infrastructure Layer**: External services, database, API clients
- **Interface Layer**: HTTP handlers, middleware, routes

## Technology Stack

- **Framework**: Gin (HTTP routing)
- **Database**: SQLite (local), PostgreSQL (production)
- **Cache**: In-memory (local), Redis (production)
- **Auth**: JWT
- **Testing**: testify/assert

## Development Roadmap

### Phase 1: Core Returns Flow (Current)
- ✅ Project setup
- ✅ Database schema
- ⏳ Core entities implementation
- ⏳ HTTP handlers
- ⏳ Returns flow logic
- ⏳ Order API integration

### Phase 2: Additional Features
- Order tracking
- Refund status
- Account management

### Phase 3: Production Readiness
- Monitoring & metrics
- Performance optimization
- Production deployment

## Contributing

1. Create a feature branch
2. Make your changes
3. Write/update tests
4. Run `make test` and `make fmt`
5. Submit a pull request

## License

Proprietary - All rights reserved

## Support

For questions or issues, please contact the development team.
