package entities

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// MessageRole represents the role of the message sender
type MessageRole string

const (
	RoleUser      MessageRole = "user"
	RoleAssistant MessageRole = "assistant"
	RoleSystem    MessageRole = "system"
)

// Message represents a message in a conversation
type Message struct {
	ID             string         `json:"id"`
	ConversationID string         `json:"conversation_id"`
	Role           MessageRole    `json:"role"`
	Content        string         `json:"content"`
	Metadata       map[string]any `json:"metadata"`
	CreatedAt      time.Time      `json:"created_at"`
}

// NewMessage creates a new message
func NewMessage(conversationID string, role MessageRole, content string) *Message {
	return &Message{
		ID:             uuid.New().String(),
		ConversationID: conversationID,
		Role:           role,
		Content:        content,
		Metadata:       make(map[string]any),
		CreatedAt:      time.Now(),
	}
}

// Validate validates the message
func (m *Message) Validate() error {
	if m.ID == "" {
		return errors.New("message ID is required")
	}
	if m.ConversationID == "" {
		return errors.New("conversation ID is required")
	}
	if m.Role != RoleUser && m.Role != RoleAssistant && m.Role != RoleSystem {
		return errors.New("invalid message role")
	}
	if m.Content == "" {
		return errors.New("message content is required")
	}
	return nil
}

// AddMetadata adds metadata to the message
func (m *Message) AddMetadata(key string, value any) {
	if m.Metadata == nil {
		m.Metadata = make(map[string]any)
	}
	m.Metadata[key] = value
}

// GetMetadata gets metadata from the message
func (m *Message) GetMetadata(key string) (any, bool) {
	if m.Metadata == nil {
		return nil, false
	}
	value, exists := m.Metadata[key]
	return value, exists
}
