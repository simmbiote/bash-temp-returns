# Swagger API Documentation

## Overview

The Customer Support API now includes comprehensive Swagger/OpenAPI documentation for all endpoints.

## Accessing Swagger UI

Once the API server is running, you can access the interactive Swagger documentation at:

```
http://localhost:8080/swagger/index.html
```

## Available Documentation Formats

The API documentation is available in multiple formats:

- **Interactive UI**: `http://localhost:8080/swagger/index.html`
- **JSON Spec**: `http://localhost:8080/swagger/doc.json`
- **YAML Spec**: Available at `/docs/swagger.yaml`

## Documented Endpoints

### Conversations

- **POST /api/v1/conversations** - Create a new conversation
  - Create a customer support conversation (natural language or guided flow)
  - Request body: `CreateConversationRequest`
  - Response: `CreateConversationResponse`

- **GET /api/v1/conversations/{id}** - Get conversation details
  - Retrieve a conversation with all its messages
  - Response: `ConversationDetailsResponse`

- **GET /api/v1/conversations/{id}/messages** - Get conversation messages
  - Retrieve all messages in a conversation
  - Response: Array of `MessageDTO`

- **POST /api/v1/conversations/{id}/messages** - Send a message
  - Send a user message and receive an AI assistant response
  - Request body: `SendMessageRequest`
  - Response: `SendMessageResponse`

### Returns

- **POST /api/v1/returns** - Create a return request
  - Create a new return request with order validation and automatic refund calculation
  - Request body: `CreateReturnRequest`
  - Response: `CreateReturnResponse`

- **GET /api/v1/returns/{id}** - Get return request details
  - Retrieve detailed information about a specific return request
  - Response: `ReturnRequestDTO`

- **GET /api/v1/returns** - List return requests
  - Retrieve all return requests for the authenticated customer
  - Query param: `customer_id` (optional, defaults to demo-customer-123)
  - Response: Array of `ReturnRequestDTO`

## Data Models

The Swagger documentation includes complete schemas for all request and response models:

### Conversation Models
- `CreateConversationRequest`
- `CreateConversationResponse`
- `ConversationDetailsResponse`
- `MessageDTO`
- `SendMessageRequest`
- `SendMessageResponse`

### Return Models
- `CreateReturnRequest`
- `ReturnItemDTO`
- `CreateReturnResponse`
- `ReturnRequestDTO`

## API Information

- **Version**: 0.2.0
- **Base URL**: `http://localhost:8080/api/v1`
- **Schemes**: HTTP, HTTPS
- **License**: MIT

## Security

The API supports Bearer token authentication (configured for future use):

```
Authorization: Bearer <token>
```

Currently, authentication is not enforced in the MVP version.

## Updating Documentation

When you make changes to the API endpoints or add new handlers, regenerate the Swagger documentation:

```bash
# Install swag CLI (if not already installed)
go install github.com/swaggo/swag/cmd/swag@latest

# Generate documentation
~/go/bin/swag init -g cmd/api/main.go -o docs
```

## Swagger Annotations

The documentation is generated from code annotations. Here's an example:

```go
// CreateConversation godoc
// @Summary Create a new conversation
// @Description Create a new customer support conversation
// @Tags Conversations
// @Accept json
// @Produce json
// @Param request body dto.CreateConversationRequest true "Request body"
// @Success 201 {object} dto.CreateConversationResponse
// @Failure 400 {object} map[string]string
// @Router /conversations [post]
func (h *ConversationHandler) CreateConversation(c *gin.Context) {
    // Handler implementation
}
```

## Testing with Swagger UI

1. Start the API server:
   ```bash
   ORDERS_API_URL=mock go run cmd/api/main.go
   ```

2. Open your browser and navigate to:
   ```
   http://localhost:8080/swagger/index.html
   ```

3. Click on any endpoint to expand it

4. Click "Try it out" to test the endpoint

5. Fill in the required parameters

6. Click "Execute" to send the request

7. View the response below

## Example: Testing Create Conversation

1. Navigate to `POST /api/v1/conversations`
2. Click "Try it out"
3. Enter the request body:
   ```json
   {
     "customer_id": "demo-customer-123",
     "type": "natural_language",
     "subject": "I need help with my order"
   }
   ```
4. Click "Execute"
5. View the response with the new conversation ID

## Example: Testing Create Return

1. Navigate to `POST /api/v1/returns`
2. Click "Try it out"
3. Enter the request body:
   ```json
   {
     "customer_id": "demo-customer-123",
     "order_number": "ORD-123456",
     "items": [
       {
         "order_item_id": "ITEM-1",
         "product_id": "PROD-456",
         "quantity": 1
       }
     ],
     "reason": "size_issue",
     "detailed_reason": "Shoes are too small",
     "refund_method": "original_payment",
     "delivery_method": "collection",
     "collection_point": "Cape Town Store"
   }
   ```
4. Click "Execute"
5. View the response with refund calculation (should be $159.99)

## Additional Features

### Export OpenAPI Spec

You can download the OpenAPI specification in JSON or YAML format:

- JSON: `http://localhost:8080/swagger/doc.json`
- YAML: Available in `/docs/swagger.yaml`

### Import to Postman

1. Download the JSON spec from `http://localhost:8080/swagger/doc.json`
2. Open Postman
3. Click "Import"
4. Select the downloaded JSON file
5. All endpoints will be imported as a collection

### Import to Insomnia

1. Download the JSON spec
2. Open Insomnia
3. Click "Import/Export"
4. Select the JSON file
5. All endpoints will be imported

## Troubleshooting

### Swagger UI not loading

- Ensure the server is running on port 8080
- Check that the `/docs` directory contains the generated files
- Verify the import: `_ "customer-support-api/docs"` is present in main.go

### Documentation not updated

- Run `swag init` command to regenerate documentation
- Restart the server after regeneration
- Clear browser cache if necessary

### Missing endpoints

- Ensure handler methods have proper Swagger annotations
- Check that routes are registered in main.go
- Regenerate documentation with `swag init`

## Benefits of Swagger Documentation

1. **Interactive Testing**: Test all endpoints directly from the browser
2. **Auto-generated**: Documentation stays in sync with code
3. **Client Generation**: Generate API clients in multiple languages
4. **Team Collaboration**: Share consistent API documentation
5. **API Contract**: Serves as a contract between frontend and backend teams

## Next Steps

- Add authentication examples to Swagger UI
- Include more detailed error response schemas
- Add request/response examples for complex scenarios
- Document query parameters and filters
- Add pagination documentation when implemented
