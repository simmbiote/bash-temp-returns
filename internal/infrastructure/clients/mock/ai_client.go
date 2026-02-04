package mock

import (
	"context"
	"strings"

	"customer-support-api/internal/domain/entities"
	"customer-support-api/internal/domain/services"
)

type mockAIClient struct{}

// NewMockAIClient creates a mock AI client for testing/development
func NewMockAIClient() services.AIClient {
	return &mockAIClient{}
}

func (c *mockAIClient) ClassifyIntent(
	ctx context.Context,
	message string,
	conversationHistory []entities.Message,
) (*services.IntentClassification, error) {
	messageLower := strings.ToLower(message)

	// Simple keyword-based classification
	if strings.Contains(messageLower, "return") {
		return &services.IntentClassification{
			Intent:     entities.IntentReturn,
			Confidence: 0.85,
			Entities:   extractOrderNumber(message),
			Reasoning:  "Message contains 'return' keyword",
		}, nil
	}

	if strings.Contains(messageLower, "refund") {
		return &services.IntentClassification{
			Intent:     entities.IntentRefund,
			Confidence: 0.80,
			Entities:   extractOrderNumber(message),
			Reasoning:  "Message contains 'refund' keyword",
		}, nil
	}

	if strings.Contains(messageLower, "track") || strings.Contains(messageLower, "where") {
		return &services.IntentClassification{
			Intent:     entities.IntentTrackOrder,
			Confidence: 0.75,
			Entities:   extractOrderNumber(message),
			Reasoning:  "Message contains tracking-related keywords",
		}, nil
	}

	if strings.Contains(messageLower, "account") || strings.Contains(messageLower, "profile") || strings.Contains(messageLower, "email") {
		return &services.IntentClassification{
			Intent:     entities.IntentAccountUpdate,
			Confidence: 0.70,
			Entities:   make(map[string]interface{}),
			Reasoning:  "Message contains account-related keywords",
		}, nil
	}

	// Default to general inquiry
	return &services.IntentClassification{
		Intent:     entities.IntentGeneralInquiry,
		Confidence: 0.60,
		Entities:   make(map[string]interface{}),
		Reasoning:  "No specific intent keywords detected",
	}, nil
}

func (c *mockAIClient) GenerateResponse(
	ctx context.Context,
	conversation *entities.Conversation,
	messages []entities.Message,
	userMessage string,
) (string, error) {
	// Generate simple rule-based response based on intent
	if conversation.Intent != nil {
		switch *conversation.Intent {
		case entities.IntentReturn:
			if strings.Contains(strings.ToLower(userMessage), "ord-") {
				return "I can help you with that return. Let me fetch the details for your order and guide you through the process.", nil
			}
			return "I'd be happy to help you with a return. Could you please provide your order number so I can look up the details?", nil

		case entities.IntentRefund:
			return "I can help you check your refund status. Could you please provide your order number or return reference?", nil

		case entities.IntentTrackOrder:
			return "I can help you track your order. Please provide your order number and I'll get the latest tracking information for you.", nil

		case entities.IntentAccountUpdate:
			return "I can assist you with updating your account details. What would you like to change?", nil
		}
	}

	// Default general response
	return "Thank you for your message. How can I assist you today? I can help with returns, refunds, order tracking, or account updates.", nil
}

// extractOrderNumber looks for order number patterns in the message
func extractOrderNumber(message string) map[string]interface{} {
	entities := make(map[string]interface{})
	messageLower := strings.ToLower(message)

	// Look for pattern like "ORD-123456" or "#123456"
	words := strings.Fields(message)
	for _, word := range words {
		wordUpper := strings.ToUpper(word)
		if strings.HasPrefix(wordUpper, "ORD-") {
			entities["order_number"] = wordUpper
			break
		}
		if strings.HasPrefix(word, "#") && len(word) > 3 {
			entities["order_number"] = "ORD-" + word[1:]
			break
		}
	}

	// Look for "order 123456" or "order number 123456"
	if strings.Contains(messageLower, "order") {
		for i, word := range words {
			if strings.Contains(strings.ToLower(word), "order") && i+1 < len(words) {
				nextWord := words[i+1]
				if strings.ToLower(nextWord) != "number" && len(nextWord) >= 3 {
					if !strings.HasPrefix(strings.ToUpper(nextWord), "ORD-") {
						entities["order_number"] = "ORD-" + nextWord
					} else {
						entities["order_number"] = strings.ToUpper(nextWord)
					}
					break
				} else if i+2 < len(words) && strings.ToLower(nextWord) == "number" {
					orderNum := words[i+2]
					if !strings.HasPrefix(strings.ToUpper(orderNum), "ORD-") {
						entities["order_number"] = "ORD-" + orderNum
					} else {
						entities["order_number"] = strings.ToUpper(orderNum)
					}
					break
				}
			}
		}
	}

	return entities
}
