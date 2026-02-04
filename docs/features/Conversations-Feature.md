# Conversations Feature

## Overview

The conversations feature enables structured customer support interactions through both natural language and guided flow types. Conversations serve as the container for customer-assistant message exchanges and track the context and intent throughout the interaction.

## Feature Scope

**Current Implementation (MVP)**:
- Conversation creation (natural language and guided flow types)
- Message creation and retrieval
- Conversation context storage
- Intent classification (structure in place, AI integration pending)
- Conversation completion tracking

**Future Enhancements**:
- AI-powered intent classification
- Natural language processing
- Multi-turn context understanding
- Guided flow step progression
- Conversation transfer to human agents
- Conversation analytics and insights

## Conversation Types

### Natural Language

**Type**: `natural_language`  
**Description**: Free-form conversation where customers express their needs in natural language and the system uses AI to classify intent and provide appropriate responses.

**Characteristics**:
- No predefined flow structure
- Intent classified from customer messages
- Context preserved across message turns
- Flexible interaction pattern
- AI-driven response generation (future)

**Use Cases**:
- General enquiries
- Complex multi-part questions
- Exploratory customer interactions
- When customer's need is unclear

### Guided Flow

**Type**: `guided_flow`  
**Description**: Structured step-by-step conversation following a predefined flow for specific tasks.

**Characteristics**:
- Defined flow with sequential steps
- `flow_id` identifies the specific flow
- `current_step` tracks progress
- Structured data collection
- Predictable interaction pattern

**Use Cases**:
- Returns processing
- Account updates
- Order tracking
- Standardised support tasks

**Required Fields**:
- `flow_id`: Identifier for the flow definition
- `current_step`: Current position in the flow

## Intent Classification

Conversations can have classified intents that guide system responses and actions.

### Intent Types

| Intent | Description | Typical Actions |
|--------|-------------|-----------------|
| `return` | Customer wants to return items | Create return request |
| `refund` | Customer enquiring about refund status | Check refund status |
| `track_order` | Customer wants order tracking info | Fetch order tracking |
| `account_update` | Customer wants to update account | Update account details |
| `general_inquiry` | General questions | Provide information |

### Intent Setting

Intents can be:
- **Manually set**: For guided flows with known purpose
- **AI-classified**: From natural language input (future)
- **Unset**: For exploratory conversations

## Use Cases

### Create Conversation

**Use Case**: `CreateConversationUseCase`  
**Location**: `internal/application/usecases/create_conversation.go`

#### Input

**DTO**: `CreateConversationRequest`

```json
{
  "type": "natural_language",
  "initial_message": "I need to return an item"
}
```

#### Processing Steps

1. **Determine Type**: Convert string type to entity type
2. **Create Conversation**: Create new conversation entity for customer
3. **Persist Conversation**: Save to database
4. **Process Initial Message** (if provided):
   - Create user message with initial content
   - Create assistant greeting response
   - Persist both messages
5. **Return Response**: Include conversation ID and assistant message

#### Output

**DTO**: `CreateConversationResponse`

```json
{
  "conversation_id": "conv-uuid",
  "type": "natural_language",
  "message": {
    "role": "assistant",
    "content": "Hello! I'm here to help you. How can I assist you today?"
  }
}
```

### Send Message

**Use Case**: `SendMessageUseCase`  
**Location**: `internal/application/usecases/send_message.go`

#### Input

**DTO**: `SendMessageRequest`

```json
{
  "content": "I want to return my order #12345",
  "metadata": {
    "platform": "mobile_app",
    "version": "2.1.0"
  }
}
```

#### Processing Steps

1. **Verify Conversation**: Confirm conversation exists
2. **Check Completion**: Prevent messages to completed conversations
3. **Create User Message**: Store customer message with metadata
4. **Process Message**: AI intent classification (placeholder in MVP)
5. **Generate Response**: Create assistant message
6. **Update Context**: Store relevant information in conversation context (future)
7. **Return Both Messages**: User and assistant messages

#### Output

**DTO**: `SendMessageResponse`

```json
{
  "user_message": {
    "id": "msg-user-uuid",
    "role": "user",
    "content": "I want to return my order #12345",
    "metadata": {"platform": "mobile_app"},
    "created_at": "2026-02-04T10:30:00Z"
  },
  "assistant_message": {
    "id": "msg-assistant-uuid",
    "role": "assistant",
    "content": "Thank you for your message. How can I assist you further?",
    "metadata": {},
    "created_at": "2026-02-04T10:30:01Z"
  }
}
```

### Get Conversation

**Use Case**: `GetConversationUseCase`  
**Location**: `internal/application/usecases/get_conversation.go`

Retrieves conversation with all messages for a customer.

**Output**: Conversation details with array of messages in chronological order.

## Context Management

Conversations maintain a flexible key-value context store to preserve information across message turns.

### Context Operations

**Set Context**:
```go
conversation.UpdateContext("order_number", "ORD-12345")
conversation.UpdateContext("return_reason", "wrong_size")
```

**Get Context**:
```go
value, exists := conversation.GetContext("order_number")
```

### Context Use Cases

- Store extracted entities (order numbers, product IDs)
- Track flow progression state
- Preserve customer preferences
- Cache API responses
- Store temporary data during multi-step processes

### Context Persistence

- Stored as JSON in database
- Survives conversation retrieval
- Cleared when conversation completes (optional)

## Message Structure

### Message Roles

**User** (`user`):
- Messages from the customer
- Input for intent classification
- Recorded for conversation history

**Assistant** (`assistant`):
- System/AI responses to customer
- Provide information and guidance
- Prompt for next steps

**System** (`system`):
- Internal system messages
- Status updates
- Action confirmations
- Not typically shown to customer

### Message Metadata

Messages support flexible metadata storage:

```json
{
  "metadata": {
    "platform": "mobile_app",
    "version": "2.1.0",
    "device_type": "ios",
    "intent_confidence": 0.95,
    "entities_extracted": ["order_number", "return_reason"]
  }
}
```

**Common Metadata Fields**:
- Platform/channel information
- Client version
- Device details
- Intent classification results
- Extracted entities
- Processing timestamps

## Conversation Lifecycle

### States

**Active**:
- `completed_at` is null
- Can receive new messages
- Context can be updated

**Completed**:
- `completed_at` is set
- Cannot receive new messages
- Read-only state
- Preserved for history

### Completion

Conversations are completed when:
- Customer ends interaction explicitly
- System task is finished (e.g., return request submitted)
- Timeout due to inactivity (future)
- Transfer to human agent (future)

**Completion Method**:
```go
conversation.Complete()
```

Sets `completed_at` timestamp and prevents further modifications.

## Data Model

### Conversation Entity

**Location**: `internal/domain/entities/conversation.go`

**Key Fields**:
- Identification: `id`, `customer_id`
- Type: `type` (natural_language or guided_flow)
- Classification: `intent`
- Flow tracking: `flow_id`, `current_step`
- Context: `context` (key-value map)
- Lifecycle: `created_at`, `updated_at`, `completed_at`

### Message Entity

**Location**: `internal/domain/entities/message.go`

**Key Fields**:
- Identification: `id`, `conversation_id`
- Content: `role`, `content`
- Additional data: `metadata` (key-value map)
- Timestamp: `created_at`

**Immutability**: Messages cannot be edited after creation.

### Database Persistence

**Tables**: `conversations`, `messages`  
**Repositories**: `ConversationRepository`, `MessageRepository`  
**Implementations**: `internal/infrastructure/persistence/sqlite/`

**Persistence Details**:
- Context stored as JSON string
- Metadata stored as JSON string
- Timestamps in RFC3339 format
- Foreign key from messages to conversations with CASCADE delete

## API Endpoints

### Create Conversation

**Endpoint**: `POST /api/v1/conversations`  
**Handler**: `ConversationHandler.CreateConversation`

**Request**:
```json
{
  "type": "natural_language",
  "initial_message": "I need help with my order"
}
```

**Response**: `201 Created`
```json
{
  "conversation_id": "conv-uuid",
  "type": "natural_language",
  "message": {
    "role": "assistant",
    "content": "Hello! I'm here to help you. How can I assist you today?"
  }
}
```

### Send Message

**Endpoint**: `POST /api/v1/conversations/:id/messages`  
**Handler**: `ConversationHandler.SendMessage`

**Request**:
```json
{
  "content": "I want to return order #12345",
  "metadata": {
    "platform": "mobile_app"
  }
}
```

**Response**: `200 OK`
```json
{
  "user_message": {
    "id": "msg-uuid-1",
    "role": "user",
    "content": "I want to return order #12345",
    "created_at": "2026-02-04T10:30:00Z"
  },
  "assistant_message": {
    "id": "msg-uuid-2",
    "role": "assistant",
    "content": "I can help you with that. Let me fetch your order details.",
    "created_at": "2026-02-04T10:30:01Z"
  }
}
```

### Get Conversation

**Endpoint**: `GET /api/v1/conversations/:id`  
**Handler**: `ConversationHandler.GetConversation`

**Response**: `200 OK`
```json
{
  "id": "conv-uuid",
  "customer_id": "cust-123",
  "type": "natural_language",
  "intent": "return",
  "context": {
    "order_number": "ORD-12345"
  },
  "messages": [
    {
      "id": "msg-1",
      "role": "user",
      "content": "I want to return order #12345",
      "created_at": "2026-02-04T10:30:00Z"
    },
    {
      "id": "msg-2",
      "role": "assistant",
      "content": "I can help you with that.",
      "created_at": "2026-02-04T10:30:01Z"
    }
  ],
  "created_at": "2026-02-04T10:30:00Z",
  "updated_at": "2026-02-04T10:30:01Z"
}
```

## Integration with Returns

Conversations can be linked to return requests:

1. Customer starts conversation about returning items
2. Intent classified as "return"
3. System creates return request linked to conversation
4. `conversation_id` stored in return request
5. Return status updates can be communicated via conversation

**Workflow**:
```
Conversation (intent: return)
    ↓ (creates)
ReturnRequest (conversation_id: conv-uuid)
```

## Future Enhancements

### Phase 2: AI Integration

- OpenAI/Azure OpenAI for intent classification
- Entity extraction from natural language
- Context-aware response generation
- Multi-turn understanding

### Phase 3: Advanced Features

- Conversation sentiment analysis
- Auto-suggestion of responses
- Multi-language support
- Voice integration

### Phase 4: Human Handoff

- Transfer to human agents
- Agent assignment
- Conversation queue management
- Response time tracking

### Phase 5: Analytics

- Conversation completion rates
- Average resolution time
- Intent distribution
- Customer satisfaction scoring
