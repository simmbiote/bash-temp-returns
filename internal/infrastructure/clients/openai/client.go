package openai

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"customer-support-api/internal/domain/entities"
	"customer-support-api/internal/domain/services"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

type openAIClient struct {
	client      *openai.Client
	model       string
	maxTokens   int64
	temperature float64
}

// NewOpenAIClient creates a new OpenAI client for production use
func NewOpenAIClient(apiKey, model string) services.AIClient {
	client := openai.NewClient(
		option.WithAPIKey(apiKey),
	)

	return &openAIClient{
		client:      client,
		model:       model,
		maxTokens:   5000,
		temperature: 0.7,
	}
}

func (c *openAIClient) ClassifyIntent(
	ctx context.Context,
	message string,
	conversationHistory []entities.Message,
) (*services.IntentClassification, error) {
	systemPrompt := `You are an intent classifier for Bash customer support. Analyse customer messages and classify intent.

Available intents:
- return: Customer wants to return items
- refund: Customer enquiring about refund status
- track_order: Customer wants order tracking
- account_update: Customer wants to update account details
- general_inquiry: General questions or unclear intent

Return ONLY valid JSON with this exact structure:
{
  "intent": "return",
  "confidence": 0.95,
  "entities": {"order_number": "ORD-123456"},
  "reasoning": "Customer explicitly mentioned wanting to return an item"
}`

	userPrompt := fmt.Sprintf("Customer message: \"%s\"\n\nClassify the intent and extract any entities (order_number, product_id, etc.).", message)

	chatCompletion, err := c.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages: openai.F([]openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(systemPrompt),
			openai.UserMessage(userPrompt),
		}),
		Model:       openai.F(c.model),
		MaxTokens:   openai.Int(int64(300)),
		Temperature: openai.Float(0.3), // Lower temperature for classification
	})

	if err != nil {
		log.Printf("OpenAI API error (classification): %v", err)
		return nil, fmt.Errorf("failed to classify intent: %w", err)
	}

	if len(chatCompletion.Choices) == 0 {
		return nil, fmt.Errorf("no response from OpenAI")
	}

	responseText := chatCompletion.Choices[0].Message.Content

	// Parse JSON response
	var result struct {
		Intent     string                 `json:"intent"`
		Confidence float64                `json:"confidence"`
		Entities   map[string]interface{} `json:"entities"`
		Reasoning  string                 `json:"reasoning"`
	}

	if err := json.Unmarshal([]byte(responseText), &result); err != nil {
		log.Printf("Failed to parse OpenAI response: %v, response: %s", err, responseText)
		return nil, fmt.Errorf("failed to parse classification response: %w", err)
	}

	// Convert string intent to entities.Intent
	intent := mapStringToIntent(result.Intent)

	return &services.IntentClassification{
		Intent:     intent,
		Confidence: result.Confidence,
		Entities:   result.Entities,
		Reasoning:  result.Reasoning,
	}, nil
}

func (c *openAIClient) GenerateResponse(
	ctx context.Context,
	conversation *entities.Conversation,
	messages []entities.Message,
	userMessage string,
) (string, error) {
	systemPrompt := `You are a helpful customer support assistant for Bash, a South African retail company.

Personality traits:
- You talk like Samuel L Jackson (casual, direct, occasional profanity)
- Friendly and professional
- Empathetic to customer concerns
- Proactive in offering solutions
- Use South African English (colour, organisation, etc.)

IMPORTANT - Conversational Flow:
- This is a CHAT API - handle everything conversationally, don't redirect to web pages
- For returns: Ask for order number if missing, then confirm item details and reason for return
- For refunds: Acknowledge concern, ask for order/return reference if missing
- For tracking: Request order number and provide status updates
- Guide customers step-by-step through the conversation
- Keep responses concise (2-3 sentences max)
- Always maintain conversation context
- Don't make promises about refund amounts or timelines
- If you have the order number, acknowledge it and move to next step`

	// Build conversation history
	historyMessages := []openai.ChatCompletionMessageParamUnion{
		openai.SystemMessage(systemPrompt),
	}

	// Add recent conversation history (last 10 messages max)
	startIdx := 0
	if len(messages) > 10 {
		startIdx = len(messages) - 10
	}

	for _, msg := range messages[startIdx:] {
		switch msg.Role {
		case entities.RoleUser:
			historyMessages = append(historyMessages, openai.UserMessage(msg.Content))
		case entities.RoleAssistant:
			historyMessages = append(historyMessages, openai.AssistantMessage(msg.Content))
		}
	}

	// Add current user message
	historyMessages = append(historyMessages, openai.UserMessage(userMessage))

	// Add intent context if available
	if conversation.Intent != nil {
		contextPrompt := fmt.Sprintf("\nDetected intent: %s", *conversation.Intent)
		historyMessages = append(historyMessages, openai.SystemMessage(contextPrompt))
	}

	chatCompletion, err := c.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages:    openai.F(historyMessages),
		Model:       openai.F(c.model),
		MaxTokens:   openai.Int(c.maxTokens),
		Temperature: openai.Float(c.temperature),
	})

	if err != nil {
		log.Printf("OpenAI API error (response generation): %v", err)
		return "", fmt.Errorf("failed to generate response: %w", err)
	}

	if len(chatCompletion.Choices) == 0 {
		return "", fmt.Errorf("no response from OpenAI")
	}

	return chatCompletion.Choices[0].Message.Content, nil
}

func mapStringToIntent(intentStr string) entities.Intent {
	switch intentStr {
	case "return":
		return entities.IntentReturn
	case "refund":
		return entities.IntentRefund
	case "track_order":
		return entities.IntentTrackOrder
	case "account_update":
		return entities.IntentAccountUpdate
	default:
		return entities.IntentGeneralInquiry
	}
}
