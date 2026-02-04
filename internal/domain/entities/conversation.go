package entities

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// ConversationType represents the type of conversation
type ConversationType string

const (
	NaturalLanguage ConversationType = "natural_language"
	GuidedFlow      ConversationType = "guided_flow"
)

// Intent represents the user's intent in the conversation
type Intent string

const (
	IntentReturn         Intent = "return"
	IntentRefund         Intent = "refund"
	IntentTrackOrder     Intent = "track_order"
	IntentAccountUpdate  Intent = "account_update"
	IntentGeneralInquiry Intent = "general_inquiry"
)

// Conversation represents a customer support conversation
type Conversation struct {
	ID          string           `json:"id"`
	CustomerID  string           `json:"customer_id"`
	Type        ConversationType `json:"type"`
	Intent      *Intent          `json:"intent,omitempty"`
	CurrentStep *string          `json:"current_step,omitempty"`
	FlowID      *string          `json:"flow_id,omitempty"`
	Context     map[string]any   `json:"context"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
	CompletedAt *time.Time       `json:"completed_at,omitempty"`
}

// NewConversation creates a new conversation
func NewConversation(customerID string, conversationType ConversationType) *Conversation {
	now := time.Now()
	return &Conversation{
		ID:         uuid.New().String(),
		CustomerID: customerID,
		Type:       conversationType,
		Context:    make(map[string]any),
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}

// Validate validates the conversation
func (c *Conversation) Validate() error {
	if c.ID == "" {
		return errors.New("conversation ID is required")
	}
	if c.CustomerID == "" {
		return errors.New("customer ID is required")
	}
	if c.Type != NaturalLanguage && c.Type != GuidedFlow {
		return errors.New("invalid conversation type")
	}
	if c.Type == GuidedFlow && c.FlowID == nil {
		return errors.New("flow ID is required for guided flow conversations")
	}
	return nil
}

// SetIntent sets the intent for the conversation
func (c *Conversation) SetIntent(intent Intent) {
	c.Intent = &intent
	c.UpdatedAt = time.Now()
}

// SetCurrentStep sets the current step in a guided flow
func (c *Conversation) SetCurrentStep(step string) {
	c.CurrentStep = &step
	c.UpdatedAt = time.Now()
}

// UpdateContext updates the conversation context
func (c *Conversation) UpdateContext(key string, value any) {
	if c.Context == nil {
		c.Context = make(map[string]any)
	}
	c.Context[key] = value
	c.UpdatedAt = time.Now()
}

// GetContext gets a value from the conversation context
func (c *Conversation) GetContext(key string) (any, bool) {
	if c.Context == nil {
		return nil, false
	}
	value, exists := c.Context[key]
	return value, exists
}

// Complete marks the conversation as completed
func (c *Conversation) Complete() {
	now := time.Now()
	c.CompletedAt = &now
	c.UpdatedAt = now
}

// IsCompleted checks if the conversation is completed
func (c *Conversation) IsCompleted() bool {
	return c.CompletedAt != nil
}
