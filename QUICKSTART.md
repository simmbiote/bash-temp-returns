# Quick Start Guide

## Prerequisites
- Go 1.21+ installed
- SQLite (comes with macOS)
- Port 8080 available

## Setup (First Time)

```bash
# 1. Navigate to project
cd /Users/simmbiote/Projects/bash-returns

# 2. Install dependencies (already done)
go mod download

# 3. Create database (already done)
make db-create

# Database is already set up at: data/chatbot.db
```

## Running the Application

```bash
# Start the server
make run

# Or run directly
go run cmd/api/main.go
```

You should see:
```
🚀 Starting Customer Support API on port 8080...
📦 Database: sqlite (./data/chatbot.db)
💾 Cache: memory
🔧 Log Level: debug

✨ API endpoints:
   GET  /health
   GET  /api/v1/ping
   POST /api/v1/conversations
   POST /api/v1/returns
```

## Testing the API

### Option 1: Using Python Script
```bash
python3 scripts/test_api.py
```

### Option 2: Using cURL

**Health Check:**
```bash
curl http://localhost:8080/health
```

**Create Conversation:**
```bash
curl -X POST http://localhost:8080/api/v1/conversations \
  -H "Content-Type: application/json" \
  -d '{
    "type": "natural_language",
    "initial_message": "I want to return my shoes"
  }'
```

**Create Return Request:**
```bash
curl -X POST http://localhost:8080/api/v1/returns \
  -H "Content-Type: application/json" \
  -d '{
    "order_number": "ORD-123456",
    "items": [
      {
        "order_item_id": "ITEM-1",
        "product_id": "PROD-456",
        "quantity": 1
      }
    ],
    "reason": "wrong_size",
    "refund_method": "original_payment",
    "delivery_method": "collect",
    "collection_point": "Store 123"
  }'
```

## Running Tests

```bash
# Run all unit tests
make test

# Run with verbose output
go test ./internal/domain/entities/... -v

# Run with coverage
go test ./internal/domain/entities/... -cover
```

Expected output:
```
16 tests PASS
coverage: 70.3% of statements
```

## Checking the Database

```bash
# Open SQLite CLI
sqlite3 data/chatbot.db

# View conversations
SELECT * FROM conversations;

# View messages
SELECT * FROM messages;

# View return requests
SELECT * FROM return_requests;

# Exit
.exit
```

## Common Commands

```bash
# Build the application
make build

# Run tests
make test

# Clean build artifacts
make clean

# Run database migrations
make db-migrate

# View all make targets
make help
```

## Troubleshooting

### Port 8080 Already in Use
```bash
# Find process using port 8080
lsof -i:8080

# Kill the process
kill -9 <PID>
```

### Database Locked
```bash
# Check for other SQLite connections
lsof data/chatbot.db

# If needed, recreate database
rm data/chatbot.db
make db-create
```

### Module Not Found
```bash
# Reinstall dependencies
go mod tidy
go mod download
```

## Project Status

✅ **Working Features:**
- Health check endpoint
- Conversation creation with initial message
- Return request creation with auto-approval
- SQLite database persistence
- JSON request/response handling
- Entity validation
- Error handling

🔄 **Coming Next (Phase 2):**
- Authentication (JWT)
- Orders API integration
- Intent classification
- Additional CRUD endpoints
- File upload for photos
- Email notifications
- Background jobs

## Need Help?

- Check [README.md](README.md) for detailed documentation
- Review [IMPLEMENTATION_SUMMARY.md](IMPLEMENTATION_SUMMARY.md) for architecture
- See [docs/bff-context.md](docs/bff-context.md) for full specification
- Run tests: `make test`
- Check logs when server is running

---

**Last Updated**: 2026-02-04
**Version**: 0.1.0 (MVP)
**Status**: ✅ Fully Functional
