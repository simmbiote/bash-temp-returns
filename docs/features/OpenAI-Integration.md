# OpenAI GPT-4o Integration

## Overview

Integration of OpenAI GPT-4o model to enable intelligent natural language processing for customer support conversations, including intent classification and contextual response generation.

**Status**: Planned  
**Target Model**: GPT-4o  
**Implementation Phase**: Phase 2

## Scope

### In Scope
- Intent classification from customer messages
- Natural language conversation handling
- Context-aware response generation
- Conversation history tracking

### Out of Scope (Separate Work)
- Guided flow conversations (manual/structured)
- Order data processing (handled by Orders API)
- Return request creation logic (handled by use cases)

## Features

### 1. Intent Classification

**Purpose**: Automatically determine customer intent from natural language input

**Current State**: 
- Intent enum defined in domain (`IntentReturn`, `IntentRefund`, `IntentTrackOrder`, `IntentAccountUpdate`, `IntentGeneralInquiry`)
- Manual setting via `conversation.SetIntent()`
- TODO comment in `SendMessageUseCase`

**Target State**:
- Automatic classification using GPT-4o
- Confidence scoring
- Multi-turn context consideration
- Entity extraction (order numbers, product IDs, etc.)

### 2. Natural Language Message Handling

**Purpose**: Generate contextual, helpful responses to customer messages

**Current State**:
- Hardcoded placeholder response: "Thank you for your message. How can I assist you further?"
- No conversation history analysis
- No context awareness

**Target State**:
- Context-aware responses using conversation history
- Proactive suggestions based on intent
- Personality consistent with Bash brand
- Handle multi-turn conversations
- Extract actionable information

## Implementation Strategy

### Phase 1: Architecture Setup (Status: Planned)

**Tasks**:
1. Define domain service interface for AI client
2. Create DTO structures for OpenAI requests/responses
3. Set up configuration management
4. Add openai-gop (openai-go) library dependency

**Deliverables**:
- `internal/domain/services/ai_client.go` - Interface definition
- `internal/application/dto/ai_dto.go` - Request/response DTOs
- Configuration in `.env` and main.go
- Updated go.mod with openai-go library

### Phase 2: Intent Classification (Status: Planned)

**Tasks**:
1. Implement OpenAI client for intent classification
2. Create prompt engineering for intent detection
3. Add confidence threshold handling
4. Integrate into `SendMessageUseCase`
5. Add fallback to `general_inquiry` for low confidence
6. Extract entities (order numbers, etc.) from messages

**Deliverables**:
- `internal/infrastructure/clients/openai/intent_classifier.go`
- Intent classification prompts
- Updated `SendMessageUseCase` with classification
- Entity extraction logic

### Phase 3: Response Generation (Status: Planned)

**Tasks**:
1. Implement conversational response generation
2. Build conversation history context
3. Create system prompts for assistant personality
4. Implement response streaming (optional)
5. Add response validation
6. Handle API errors gracefully

**Deliverables**:
- `internal/infrastructure/clients/openai/response_generator.go`
- System prompts for brand voice
- Updated `SendMessageUseCase` with GPT-4o responses
- Error handling and fallbacks

### Phase 4: Testing & Refinement (Status: Planned)

**Tasks**:
1. Unit tests with mocked OpenAI responses
2. Integration tests with real API (test mode)
3. Prompt optimisation based on results
4. Performance monitoring
5. Cost tracking implementation
6. Rate limiting

**Deliverables**:
- Comprehensive test suite
- Optimised prompts
- Performance metrics
- Cost monitoring dashboard

## Technical Design

### Domain Service Interface

**Location**: `internal/domain/services/ai_client.go`

```go
type AIClient interface {
    // ClassifyIntent analyses message content and returns intent with confidence
    ClassifyIntent(ctx context.Context, message string, conversationHistory []Message) (*IntentClassification, error)
    
    // GenerateResponse creates contextual assistant response
    GenerateResponse(ctx context.Context, conversation *Conversation, messages []Message) (string, error)
    
    // ExtractEntities pulls structured data from natural language
    ExtractEntities(ctx context.Context, message string) (map[string]interface{}, error)
}

type IntentClassification struct {
    Intent     Intent
    Confidence float64
    Entities   map[string]interface{}
    Reasoning  string
}
```

### Infrastructure Implementation

**Location**: `internal/infrastructure/clients/openai/`

**Structure**:
```
openai/
├── client.go              # Main OpenAI client
├── intent_classifier.go   # Intent classification logic
├── response_generator.go  # Response generation logic
├── prompts.go            # Prompt templates
└── config.go             # OpenAI configuration
```

### Configuration

**Environment Variables**:
```bash
OPENAI_API_KEY=sk-...
OPENAI_MODEL=gpt-4o
OPENAI_MAX_TOKENS=500
OPENAI_TEMPERATURE=0.7
AI_PROVIDER=openai  # or 'mock' for testing
```

### Integration Points

**SendMessageUseCase** will be updated:
```go
func (uc *SendMessageUseCase) Execute(...) (*dto.SendMessageResponse, error) {
    // 1. Save user message (existing)
    
    // 2. Classify intent using AI (NEW)
    classification, err := uc.aiClient.ClassifyIntent(ctx, req.Content, conversationHistory)
    if err != nil {
        // Fallback: use general_inquiry
    }
    
    // 3. Update conversation intent if classified (NEW)
    if classification.Confidence > 0.8 {
        conversation.SetIntent(classification.Intent)
        uc.convRepo.Update(ctx, conversation)
    }
    
    // 4. Generate contextual response (NEW)
    responseContent, err := uc.aiClient.GenerateResponse(ctx, conversation, messages)
    if err != nil {
        // Fallback: use default message
    }
    
    // 5. Save assistant message (existing, updated content)
    
    // 6. Return both messages (existing)
}
```

## Prompt Engineering

### Intent Classification Prompt

**System Prompt**:
```
You are an intent classifier for Bash customer support. Analyse customer messages and classify intent.

Available intents:
- return: Customer wants to return items
- refund: Customer enquiring about refund status
- track_order: Customer wants order tracking
- account_update: Customer wants to update account details
- general_inquiry: General questions or unclear intent

Return JSON with: intent, confidence (0-1), entities (order_number, product_id, etc.)
```

**User Prompt Template**:
```
Customer message: "{message}"

Previous conversation context:
{conversation_history}

Classify the intent and extract any entities.
```

### Response Generation Prompt

**System Prompt**:
```
You are a helpful customer support assistant for Bash, a South African retail company.

Personality traits:
- Friendly and professional
- Empathetic to customer concerns
- Proactive in offering solutions
- Use South African English (colour, organisation, etc.)

Guidelines:
- Keep responses concise (2-3 sentences)
- If customer wants to return, guide them to provide order number
- If order number detected, confirm and proceed
- Always maintain conversation context
- Don't make promises about refund amounts or timelines
```

**User Prompt Template**:
```
Conversation history:
{messages}

Current customer message: "{message}"
Detected intent: {intent}
Extracted entities: {entities}

Generate a helpful, contextual response.
```

## Error Handling

### API Failures

**Strategy**: Graceful degradation

```go
response, err := aiClient.GenerateResponse(ctx, conversation, messages)
if err != nil {
    log.Errorf("OpenAI API error: %v", err)
    // Fallback to default response
    response = "I'm here to help. Could you tell me more about what you need?"
}
```

### Rate Limiting

**Approach**: Exponential backoff with circuit breaker

**Configuration**:
- Max retries: 3
- Initial backoff: 1s
- Max backoff: 10s
- Circuit breaker threshold: 5 consecutive failures

### Cost Management

**Monitoring**:
- Track token usage per request
- Log daily/monthly costs
- Alert on budget thresholds
- Implement request throttling if needed

## Testing Strategy

### Unit Tests

**Mock OpenAI Client**:
```go
type mockAIClient struct {
    intentResponse *IntentClassification
    responseText   string
    err           error
}
```

**Test Cases**:
- Intent classification with high confidence
- Intent classification with low confidence (fallback)
- Response generation with context
- Error handling and fallbacks
- Entity extraction

### Integration Tests

**Approach**: Use OpenAI test mode or separate test account

**Test Scenarios**:
- Real API calls with sample conversations
- Prompt effectiveness validation
- Response quality assessment
- Performance benchmarks

## Performance Considerations

### Latency

**Target**: < 2 seconds for response generation

**Optimisations**:
- Cache common responses (future)
- Stream responses (optional)
- Parallel processing where possible
- Use shorter context windows when appropriate

### Token Usage

**Strategy**: Minimise costs while maintaining quality

**Techniques**:
- Limit conversation history to last N messages
- Summarise older context (future)
- Use appropriate max_tokens settings
- Monitor and optimise prompts

## Migration Path

### Current to Phase 1
1. Add openai-go library to go.mod (`go get github.com/openai/openai-go`)
2. Create domain interface
3. Update SendMessageUseCase signature to accept AIClient
4. Add mock implementation for testing
5. Update dependency injection in main.go

### Phase 1 to Phase 2
1. Implement intent classifier
2. Test with real API
3. Integrate into SendMessageUseCase
4. Deploy with feature flag
5. Monitor accuracy

### Phase 2 to Phase 3
1. Implement response generator
2. Optimise prompts
3. Replace hardcoded responses
4. A/B test responses (optional)
5. Full rollout

## Security Considerations

- API key stored in environment variables (never in code)
- Sanitise user input before sending to OpenAI
- Don't send sensitive customer data unnecessarily
- Implement request logging for audit
- Rate limiting per customer
- Monitor for abuse/spam

## Dependencies

**New Packages**:
```go
require (
    github.com/openai/openai-go v0.1.0-alpha.38
)
```

**Library**: Using official `openai-gop` (openai-go) library from OpenAI

**Configuration Dependencies**:
- `OPENAI_API_KEY` environment variable
- Valid OpenAI account with API access
- Billing setup for API usage

## Future Enhancements

### Phase 5+
- Multi-language support
- Conversation summarisation
- Sentiment analysis
- Proactive problem detection
- Custom fine-tuned model for Bash-specific queries
- Voice integration (speech-to-text)

## Related Documentation

- [Conversations Feature](Conversations-Feature.md) - Context on conversation types and intents
- [Architecture and Design Principles](../architecture/Architecture-and-Design-Principles.md) - Service interface patterns
- [Testing Strategy](../testing/Testing-Strategy.md) - Testing approach for AI integration

## Approval Required

Before proceeding with implementation:
- [ ] Confirm OpenAI account and API key access
- [ ] Confirm budget for API usage
- [ ] Review and approve prompt templates
- [ ] Approve phased rollout plan
- [ ] Confirm test data requirements
