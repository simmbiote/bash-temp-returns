package services

import (
	"context"
	"customer-support-api/internal/domain/entities"
)

// AIClient defines the interface for AI-powered conversation processing
type AIClient interface {
	// ClassifyIntent analyses message content and returns intent with confidence
	ClassifyIntent(ctx context.Context, message string, conversationHistory []entities.Message) (*IntentClassification, error)

	// GenerateResponse creates contextual assistant response based on conversation
	GenerateResponse(ctx context.Context, conversation *entities.Conversation, messages []entities.Message, userMessage string) (string, error)
}

// IntentClassification represents the result of intent classification
type IntentClassification struct {
	Intent     entities.Intent
	Confidence float64
	Entities   map[string]interface{}
	Reasoning  string
}
