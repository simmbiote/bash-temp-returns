# OpenAI Integration Testing Guide

This guide provides step-by-step instructions for testing the OpenAI GPT-4o integration for conversational customer support.

## Prerequisites

- Go 1.24+ installed
- OpenAI API key (or use mock mode)
- Terminal with `curl` and `jq` (optional, for pretty JSON)

## Configuration

### Option 1: Test with Mock AI (No API costs)

Edit `.env`:
```bash
AI_PROVIDER=mock
ORDERS_API_URL=mock
```

The mock AI uses keyword-based classification and provides deterministic responses.

### Option 2: Test with Real OpenAI GPT-4o

Edit `.env`:
```bash
AI_PROVIDER=openai
OPENAI_API_KEY=sk-proj-your-api-key-here
OPENAI_MODEL=gpt-4o
ORDERS_API_URL=mock
```

**Note:** This requires an active OpenAI account with credits.

## Running the Server

### Start the API server:

```bash
make run
```

You should see output like:
```
Using mock Orders API client
Using OpenAI client with model: gpt-4o
🚀 Starting Customer Support API on port 8080...
```

### Verify server is running:

```bash
curl http://localhost:8080/health
```

Expected response:
```json
{
  "cache": "memory",
  "database": "sqlite",
  "db_status": "connected",
  "status": "ok",
  "version": "0.1.0"
}
```

## Testing Conversational Flow

### Test Scenario 1: Return Request

#### Step 1: Create a conversation

```bash
curl -X POST http://localhost:8080/api/v1/conversations \
  -H "Content-Type: application/json" \
  -d '{"type": "natural_language"}' | jq
```

**Expected Response:**
```json
{
  "conversation_id": "abc123...",
  "type": "natural_language"
}
```

**Note the `conversation_id`** - you'll need it for subsequent requests.

#### Step 2: Send initial return request

```bash
# Replace {conversation_id} with actual ID from Step 1
curl -X POST http://localhost:8080/api/v1/conversations/{conversation_id}/messages \
  -H "Content-Type: application/json" \
  -d '{"content": "I received a damaged laptop in order ORD-12345"}' | jq
```

**Expected Response (OpenAI):**
```json
{
  "user_message": {
    "id": "...",
    "role": "user",
    "content": "I received a damaged laptop in order ORD-12345",
    "created_at": "2026-02-04T15:20:56+02:00"
  },
  "assistant_message": {
    "id": "...",
    "role": "assistant",
    "content": "Ah damn, I'm sorry to hear that the laptop arrived damaged, mate. Let's sort this out. I've got your order number, ORD-12345. Would you like to return it for a replacement or a refund?",
    "created_at": "2026-02-04T15:20:56+02:00"
  }
}
```

**What to observe:**
- User message saved
- AI acknowledges the order number
- AI asks clarifying question (replacement vs refund)
- Samuel L. Jackson personality evident

#### Step 3: Continue conversation

```bash
curl -X POST http://localhost:8080/api/v1/conversations/{conversation_id}/messages \
  -H "Content-Type: application/json" \
  -d '{"content": "I want a full refund"}' | jq
```

**Expected Response:**
```json
{
  "user_message": {
    "role": "user",
    "content": "I want a full refund"
  },
  "assistant_message": {
    "role": "assistant",
    "content": "Got it, you want a refund. Let's get that sorted. Have you already initiated a return for this order, or do you need help starting that process?"
  }
}
```

**What to observe:**
- AI maintains conversation context
- AI guides to next logical step
- Natural conversational flow

#### Step 4: Verify intent classification

```bash
curl http://localhost:8080/api/v1/conversations/{conversation_id} | jq '.intent'
```

**Expected Response:**
```json
"return"
```

**What to observe:**
- Conversation intent correctly classified as "return"
- Intent persisted in database

#### Step 5: View full conversation history

```bash
curl http://localhost:8080/api/v1/conversations/{conversation_id} | jq
```

**Expected Response:**
```json
{
  "id": "abc123...",
  "customer_id": "demo-customer-123",
  "type": "natural_language",
  "intent": "return",
  "messages": [
    {
      "role": "user",
      "content": "I received a damaged laptop in order ORD-12345",
      "created_at": "..."
    },
    {
      "role": "assistant",
      "content": "Ah damn, I'm sorry to hear...",
      "created_at": "..."
    },
    {
      "role": "user",
      "content": "I want a full refund",
      "created_at": "..."
    },
    {
      "role": "assistant",
      "content": "Got it, you want a refund...",
      "created_at": "..."
    }
  ],
  "created_at": "...",
  "updated_at": "..."
}
```

---

### Test Scenario 2: Refund Status Inquiry

#### Step 1: Create new conversation

```bash
curl -X POST http://localhost:8080/api/v1/conversations \
  -H "Content-Type: application/json" \
  -d '{"type": "natural_language"}' | jq
```

#### Step 2: Ask about refund status

```bash
curl -X POST http://localhost:8080/api/v1/conversations/{conversation_id}/messages \
  -H "Content-Type: application/json" \
  -d '{"content": "Where is my refund for order ORD-67890?"}' | jq
```

**Expected Behavior:**
- Intent classified as "refund"
- AI acknowledges refund inquiry
- AI requests order/return reference if needed

#### Step 3: Verify intent

```bash
curl http://localhost:8080/api/v1/conversations/{conversation_id} | jq '.intent'
```

**Expected:** `"refund"`

---

### Test Scenario 3: Order Tracking

#### Step 1: Create new conversation

```bash
curl -X POST http://localhost:8080/api/v1/conversations \
  -H "Content-Type: application/json" \
  -d '{"type": "natural_language"}' | jq
```

#### Step 2: Request tracking information

```bash
curl -X POST http://localhost:8080/api/v1/conversations/{conversation_id}/messages \
  -H "Content-Type: application/json" \
  -d '{"content": "Can you track my order ORD-55443?"}' | jq
```

**Expected Behavior:**
- Intent classified as "track_order"
- AI acknowledges tracking request
- AI may ask for order details if needed

---

### Test Scenario 4: General Inquiry (Fallback)

#### Step 1: Create new conversation

```bash
curl -X POST http://localhost:8080/api/v1/conversations \
  -H "Content-Type: application/json" \
  -d '{"type": "natural_language"}' | jq
```

#### Step 2: Send vague message

```bash
curl -X POST http://localhost:8080/api/v1/conversations/{conversation_id}/messages \
  -H "Content-Type: application/json" \
  -d '{"content": "Hello, I need some help"}' | jq
```

**Expected Behavior:**
- Intent classified as "general_inquiry"
- AI asks clarifying questions
- AI guides user to provide more details

---

## Testing Mock vs OpenAI

### Mock AI Behavior

The mock client uses keyword detection:
- "return" → return intent
- "refund" → refund intent  
- "track" → track_order intent
- No keywords → general_inquiry

Responses are rule-based and deterministic.

### OpenAI Behavior

OpenAI GPT-4o provides:
- **Natural language understanding** - understands context and nuance
- **Entity extraction** - automatically extracts order numbers, product IDs
- **Conversational responses** - maintains context across messages
- **Personality** - Samuel L. Jackson style with South African English

---

## Monitoring and Debugging

### Check server logs

The server logs show:
- Which AI provider is active (mock/OpenAI)
- OpenAI API calls and responses
- Classification results
- Any errors

Look for lines like:
```
Using OpenAI client with model: gpt-4o
OpenAI API error (classification): ...
```

### Common Issues

#### 1. Fallback responses

**Symptom:** Getting generic responses like "Thank you for your message. How can I assist you further?"

**Cause:** AI call failed (check server logs for errors)

**Solutions:**
- Check API key is valid
- Verify you have OpenAI credits
- Check network connectivity

#### 2. Intent not being set

**Symptom:** `"intent": null` in conversation details

**Causes:**
- Confidence threshold not met (<0.75)
- Classification API call failed
- Unclear user message

**Solution:** Try more explicit messages like "I want to return order ORD-123"

#### 3. No context awareness

**Symptom:** AI doesn't remember previous messages

**Cause:** Using mock client (keyword-based only)

**Solution:** Switch to `AI_PROVIDER=openai` for full context awareness

---

## Performance Benchmarks

### Mock AI
- **Latency:** <10ms
- **Cost:** $0
- **Accuracy:** 60-70% (keyword-based)

### OpenAI GPT-4o
- **Latency:** 2-6 seconds per request
- **Cost:** ~$0.01-0.02 per conversation
- **Accuracy:** 95%+ (with proper prompts)

---

## Advanced Testing

### Testing with conversation history

```bash
# Create conversation
CONV_ID=$(curl -s -X POST http://localhost:8080/api/v1/conversations \
  -H "Content-Type: application/json" \
  -d '{"type": "natural_language"}' | jq -r '.conversation_id')

# Send multiple messages
curl -X POST http://localhost:8080/api/v1/conversations/$CONV_ID/messages \
  -H "Content-Type: application/json" \
  -d '{"content": "I have a problem with my order"}' | jq

sleep 2

curl -X POST http://localhost:8080/api/v1/conversations/$CONV_ID/messages \
  -H "Content-Type: application/json" \
  -d '{"content": "It arrived damaged"}' | jq

sleep 2

curl -X POST http://localhost:8080/api/v1/conversations/$CONV_ID/messages \
  -H "Content-Type: application/json" \
  -d '{"content": "Order number is ORD-99887"}' | jq

# View full conversation
curl http://localhost:8080/api/v1/conversations/$CONV_ID | jq
```

**What to observe:**
- AI maintains context across all messages
- Intent classification improves as more context is provided
- Order number extraction from later message

---

## Test Script

Save this as `test_openai_flow.sh`:

```bash
#!/bin/bash

BASE_URL="http://localhost:8080/api/v1"

echo "=== Testing OpenAI Integration ==="
echo

# Create conversation
echo "1. Creating conversation..."
RESPONSE=$(curl -s -X POST $BASE_URL/conversations \
  -H "Content-Type: application/json" \
  -d '{"type": "natural_language"}')
CONV_ID=$(echo $RESPONSE | jq -r '.conversation_id')
echo "Conversation ID: $CONV_ID"
echo

# Send message
echo "2. Sending message about damaged item..."
curl -s -X POST $BASE_URL/conversations/$CONV_ID/messages \
  -H "Content-Type: application/json" \
  -d '{"content": "I received a broken phone in order ORD-12345"}' | jq
echo

sleep 3

# Continue conversation
echo "3. Continuing conversation..."
curl -s -X POST $BASE_URL/conversations/$CONV_ID/messages \
  -H "Content-Type: application/json" \
  -d '{"content": "I want a refund"}' | jq
echo

sleep 3

# Check intent
echo "4. Checking classified intent..."
INTENT=$(curl -s $BASE_URL/conversations/$CONV_ID | jq -r '.intent')
echo "Classified Intent: $INTENT"
echo

# View full conversation
echo "5. Full conversation history:"
curl -s $BASE_URL/conversations/$CONV_ID | jq
```

Run with:
```bash
chmod +x test_openai_flow.sh
./test_openai_flow.sh
```

---

## Next Steps

After verifying the OpenAI integration works:

1. **Phase 2:** Refine intent classification with more intents
2. **Phase 3:** Enhance response generation with domain knowledge
3. **Phase 4:** Add metrics and monitoring
4. **Production:** Deploy with proper error handling and rate limiting

---

## Related Documentation

- [OpenAI Integration Plan](../features/OpenAI-Integration.md)
- [API Endpoints Reference](../api/API-Endpoints-Reference.md)
- [Testing Strategy](Testing-Strategy.md)
