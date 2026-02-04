# Customer Support Chatbot API - Overview

## Purpose

The Customer Support Chatbot API is a Go-based REST API that powers self-service customer support functionality for Bash retail operations. The system enables customers to perform support tasks through both natural language interactions and structured guided flows, with primary focus on returns processing.

## Project Status

**Version**: 0.1.0 (MVP)  
**Date**: February 2026  
**Status**: Development

## Core Capabilities

- **Returns Processing**: Complete returns workflow from order selection through refund processing
- **Conversational Interface**: Natural language and guided flow conversation types
- **Order Integration**: Integration with Bash Orders API for order validation and data retrieval
- **Persistent Storage**: SQLite-based data persistence for conversations, messages, and return requests
- **RESTful API**: HTTP endpoints following REST principles

## Technology Stack

- **Language**: Go 1.24
- **Web Framework**: Gin v1.11.0
- **Database**: SQLite v1.44.3 (modernc.org/sqlite - pure Go implementation)
- **Authentication**: JWT v5.3.1 (planned for future implementation)
- **Identifiers**: UUID v1.6.0

## Architecture

The project follows **Clean Architecture** principles with clear separation of concerns:

```
├── cmd/api/              # Application entry point
├── internal/
│   ├── domain/          # Core business logic (entities, repositories, services)
│   ├── application/     # Use cases and DTOs
│   ├── infrastructure/  # External integrations (database, APIs, clients)
│   └── interfaces/      # HTTP handlers and routing
├── migrations/          # Database schema migrations
└── docs/               # Technical documentation
```

See [Architecture and Design Principles](architecture/Architecture-and-Design-Principles.md) for detailed architectural decisions.

## Key Features

### Returns Processing
Complete self-service returns workflow with status tracking, multiple refund methods, and delivery options. See [Returns Feature](features/Returns-Feature.md).

### Conversation Management
Support for both natural language and guided flow conversations with intent classification and context tracking. See [Conversations Feature](features/Conversations-Feature.md).

### Order Integration
Integration with Bash Orders API for order validation and retrieval. See [Orders API Integration](integrations/Orders-API-Integration.md).

## Domain Model

Core entities:
- **Conversation**: Represents a customer support interaction session
- **Message**: Individual messages within a conversation
- **ReturnRequest**: Return request with items, status workflow, and refund details

See [Domain Model](architecture/Domain-Model.md) for complete entity specifications.

## Data Persistence

SQLite database with 7 tables supporting conversations, messages, return requests, refund requests, account actions, flows, and audit logs.

See [Database Schema](architecture/Database-Schema.md) for schema details.

## API Endpoints

RESTful HTTP endpoints for conversation and return request management:
- Health check and status endpoints
- Conversation creation and message handling
- Return request creation and retrieval

See [API Endpoints Reference](api/API-Endpoints-Reference.md) for complete API specification.

## Testing

The project includes unit tests with 70.3% code coverage for domain entities. Integration and end-to-end testing infrastructure is in place.

See [Testing Strategy](testing/Testing-Strategy.md) for testing approach and coverage.

## Getting Started

### Prerequisites
- Go 1.21 or higher
- SQLite3
- Port 8080 available

### Quick Start

```bash
# Navigate to project
cd /Users/andriesd/Documents/Bash/bash-temp-returns

# Install dependencies
make deps

# Create database
make db-create

# Run application
make run
```

The server will start on `http://localhost:8080`

### Verify Installation

```bash
curl http://localhost:8080/health
```

Expected response:
```json
{
  "status": "ok",
  "database": "sqlite",
  "db_status": "connected",
  "cache": "memory",
  "version": "0.1.0"
}
```

## Development Commands

```bash
make help              # Show all available commands
make build             # Build the application
make run               # Run the application
make test              # Run all tests
make test-coverage     # Run tests with coverage report
make db-create         # Create and initialise database
make db-reset          # Reset database
make clean             # Clean build artifacts
make fmt               # Format code
```

## MVP Scope

Current implementation includes:
- Basic conversation creation and retrieval
- Complete returns workflow (auto-approval in MVP)
- Mock Orders API client for development
- SQLite persistence layer
- Unit test coverage for domain entities

## Future Enhancements

Planned features beyond MVP:
- JWT-based customer authentication
- Real Orders API integration
- AI-powered intent classification
- Refund request approval workflow
- Account management actions
- Enhanced guided flow capabilities

## Documentation Structure

```
docs/
├── Overview.md (this file)
├── architecture/
│   ├── Architecture-and-Design-Principles.md
│   ├── Domain-Model.md
│   └── Database-Schema.md
├── features/
│   ├── Returns-Feature.md
│   └── Conversations-Feature.md
├── integrations/
│   └── Orders-API-Integration.md
├── api/
│   └── API-Endpoints-Reference.md
└── testing/
    └── Testing-Strategy.md
```

## Related Documentation

- [QUICKSTART.md](../QUICKSTART.md) - Detailed setup guide
- [README.md](../README.md) - Project readme
- [IMPLEMENTATION_SUMMARY.md](../IMPLEMENTATION_SUMMARY.md) - Implementation checklist
- [bff-context.md](bff-context.md) - Original BFF integration specification
- [go-chatbot-api-spec.md](go-chatbot-api-spec.md) - API specification
