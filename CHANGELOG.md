# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.0] - 2026-02-04

### Added
- Initial project setup with Clean Architecture structure
- Go module initialization (customer-support-api)
- SQLite database schema with 7 tables
- Domain entities: Conversation, Message, ReturnRequest
- Repository interfaces for all entities
- SQLite repository implementations with JSON marshaling
- DTOs for API requests/responses
- Use cases: CreateConversation, CreateReturnRequest
- HTTP handlers for conversations and returns
- Main API server with Gin framework
- Health check endpoint
- Comprehensive unit tests (16 tests, 70.3% coverage)
- Makefile with common commands
- Environment configuration (.env)
- Documentation (README, QUICKSTART, IMPLEMENTATION_SUMMARY)
- Test scripts (Python and Bash)

### Features
- **POST /api/v1/conversations** - Create new conversation with initial message
- **POST /api/v1/returns** - Create and submit return request
- **GET /health** - Health check with database status
- **GET /api/v1/ping** - Simple ping endpoint

### Technical Details
- Clean Architecture with dependency inversion
- SQLite for development (easy PostgreSQL migration later)
- JSON marshaling for complex types (context, metadata, items, photos)
- RFC3339 timestamps
- Comprehensive error handling with custom AppError type
- Nullable field handling with sql.NullString
- Auto-approval workflow for returns (MVP)
- Demo customer ID for authentication (JWT planned for Phase 2)

### Known Issues
- No authentication yet (using demo customer ID)
- Refund amounts not calculated (set to 0)
- No Orders API integration
- No intent classification
- No email notifications
- No file upload for photos
- No pagination for list endpoints

### Dependencies
- github.com/gin-gonic/gin v1.11.0
- modernc.org/sqlite v1.44.3
- github.com/google/uuid v1.6.0
- github.com/golang-jwt/jwt/v5 v5.3.1
- github.com/joho/godotenv v1.5.1

### Testing
- 16 unit tests passing
- 70.3% code coverage on domain layer
- Manual API integration tests successful
- Database persistence verified

## [Unreleased]

### Planned for Phase 2
- JWT authentication middleware
- Additional CRUD endpoints (GET conversation, list conversations, etc.)
- Orders API client integration
- AI-based intent classification
- Guided flow engine
- Background job processing
- Email notifications
- File upload for return photos
- Pagination and filtering
- Enhanced error handling and logging
- Integration tests
- E2E tests
- PostgreSQL support
- Docker containerization
- CI/CD pipeline

---

[0.1.0]: https://github.com/yourusername/bash-returns/releases/tag/v0.1.0
