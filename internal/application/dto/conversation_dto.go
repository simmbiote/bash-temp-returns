package dto

type CreateConversationRequest struct {
	Type           string `json:"type" binding:"required,oneof=natural_language guided_flow"`
	InitialMessage string `json:"initial_message,omitempty"`
}

type CreateConversationResponse struct {
	ConversationID string      `json:"conversation_id"`
	Type           string      `json:"type"`
	Intent         *string     `json:"intent,omitempty"`
	Message        *MessageDTO `json:"message,omitempty"`
}

type SendMessageRequest struct {
	Content  string                 `json:"content" binding:"required"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

type SendMessageResponse struct {
	UserMessage      MessageDTO `json:"user_message"`
	AssistantMessage MessageDTO `json:"assistant_message"`
}

type MessageDTO struct {
	ID        string                 `json:"id"`
	Role      string                 `json:"role"`
	Content   string                 `json:"content"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt string                 `json:"created_at"`
}

type ConversationDetailsResponse struct {
	ID          string                 `json:"id"`
	CustomerID  string                 `json:"customer_id"`
	Type        string                 `json:"type"`
	Intent      *string                `json:"intent,omitempty"`
	CurrentStep *string                `json:"current_step,omitempty"`
	FlowID      *string                `json:"flow_id,omitempty"`
	Context     map[string]interface{} `json:"context"`
	Messages    []MessageDTO           `json:"messages"`
	CreatedAt   string                 `json:"created_at"`
	UpdatedAt   string                 `json:"updated_at"`
	CompletedAt *string                `json:"completed_at,omitempty"`
}

type ConversationStateDTO struct {
	CurrentStep string                 `json:"current_step,omitempty"`
	Data        map[string]interface{} `json:"data,omitempty"`
}
