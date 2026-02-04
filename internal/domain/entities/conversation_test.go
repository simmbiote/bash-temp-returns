package entities

import (
	"testing"
	"time"
)

func TestNewConversation(t *testing.T) {
	customerID := "cust-123"
	conv := NewConversation(customerID, NaturalLanguage)

	if conv.ID == "" {
		t.Error("Expected ID to be generated")
	}
	if conv.CustomerID != customerID {
		t.Errorf("Expected CustomerID to be %s, got %s", customerID, conv.CustomerID)
	}
	if conv.Type != NaturalLanguage {
		t.Errorf("Expected Type to be %s, got %s", NaturalLanguage, conv.Type)
	}
	if conv.Context == nil {
		t.Error("Expected Context to be initialized")
	}
	if conv.CreatedAt.IsZero() {
		t.Error("Expected CreatedAt to be set")
	}
	if conv.UpdatedAt.IsZero() {
		t.Error("Expected UpdatedAt to be set")
	}
}

func TestConversationValidate(t *testing.T) {
	tests := []struct {
		name    string
		conv    *Conversation
		wantErr bool
	}{
		{
			name: "valid natural language conversation",
			conv: &Conversation{
				ID:         "conv-1",
				CustomerID: "cust-1",
				Type:       NaturalLanguage,
				Context:    make(map[string]any),
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
			},
			wantErr: false,
		},
		{
			name: "valid guided flow conversation",
			conv: func() *Conversation {
				flowID := "flow-1"
				return &Conversation{
					ID:         "conv-2",
					CustomerID: "cust-1",
					Type:       GuidedFlow,
					FlowID:     &flowID,
					Context:    make(map[string]any),
					CreatedAt:  time.Now(),
					UpdatedAt:  time.Now(),
				}
			}(),
			wantErr: false,
		},
		{
			name: "missing ID",
			conv: &Conversation{
				CustomerID: "cust-1",
				Type:       NaturalLanguage,
			},
			wantErr: true,
		},
		{
			name: "missing customer ID",
			conv: &Conversation{
				ID:   "conv-1",
				Type: NaturalLanguage,
			},
			wantErr: true,
		},
		{
			name: "invalid type",
			conv: &Conversation{
				ID:         "conv-1",
				CustomerID: "cust-1",
				Type:       "invalid",
			},
			wantErr: true,
		},
		{
			name: "guided flow without flow ID",
			conv: &Conversation{
				ID:         "conv-1",
				CustomerID: "cust-1",
				Type:       GuidedFlow,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.conv.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConversationSetIntent(t *testing.T) {
	conv := NewConversation("cust-1", NaturalLanguage)
	oldUpdatedAt := conv.UpdatedAt
	time.Sleep(10 * time.Millisecond)

	conv.SetIntent(IntentReturn)

	if conv.Intent == nil {
		t.Error("Expected Intent to be set")
	}
	if *conv.Intent != IntentReturn {
		t.Errorf("Expected Intent to be %s, got %s", IntentReturn, *conv.Intent)
	}
	if !conv.UpdatedAt.After(oldUpdatedAt) {
		t.Error("Expected UpdatedAt to be updated")
	}
}

func TestConversationContext(t *testing.T) {
	conv := NewConversation("cust-1", NaturalLanguage)

	// Test UpdateContext
	conv.UpdateContext("order_number", "ORD-123")
	value, exists := conv.GetContext("order_number")
	if !exists {
		t.Error("Expected context value to exist")
	}
	if value != "ORD-123" {
		t.Errorf("Expected context value to be ORD-123, got %v", value)
	}

	// Test non-existent key
	_, exists = conv.GetContext("non_existent")
	if exists {
		t.Error("Expected non-existent key to return false")
	}
}

func TestConversationComplete(t *testing.T) {
	conv := NewConversation("cust-1", NaturalLanguage)

	if conv.IsCompleted() {
		t.Error("Expected conversation to not be completed initially")
	}

	conv.Complete()

	if !conv.IsCompleted() {
		t.Error("Expected conversation to be completed")
	}
	if conv.CompletedAt == nil {
		t.Error("Expected CompletedAt to be set")
	}
}
