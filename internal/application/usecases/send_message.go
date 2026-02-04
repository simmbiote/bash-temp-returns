package usecases

import (
	"context"
	"customer-support-api/internal/application/dto"
	"customer-support-api/internal/domain/entities"
	"customer-support-api/internal/domain/repositories"
	"customer-support-api/internal/pkg/errors"
)

type SendMessageUseCase struct {
	convRepo    repositories.ConversationRepository
	messageRepo repositories.MessageRepository
}

func NewSendMessageUseCase(
	convRepo repositories.ConversationRepository,
	messageRepo repositories.MessageRepository,
) *SendMessageUseCase {
	return &SendMessageUseCase{
		convRepo:    convRepo,
		messageRepo: messageRepo,
	}
}

func (uc *SendMessageUseCase) Execute(
	ctx context.Context,
	conversationID string,
	req dto.SendMessageRequest,
) (*dto.SendMessageResponse, error) {
	// Verify conversation exists
	conversation, err := uc.convRepo.GetByID(ctx, conversationID)
	if err != nil {
		return nil, err
	}

	// Check if conversation is completed
	if conversation.IsCompleted() {
		return nil, errors.NewValidationError("conversation is already completed", nil)
	}

	// Create user message
	userMessage := entities.NewMessage(conversationID, entities.RoleUser, req.Content)
	if req.Metadata != nil {
		for key, value := range req.Metadata {
			userMessage.AddMetadata(key, value)
		}
	}

	if err := uc.messageRepo.Create(ctx, userMessage); err != nil {
		return nil, errors.NewDatabaseError("failed to save user message", err)
	}

	// TODO: Process message with AI/intent classification
	// For now, create a simple assistant response
	assistantMessage := entities.NewMessage(
		conversationID,
		entities.RoleAssistant,
		"Thank you for your message. How can I assist you further?",
	)

	if err := uc.messageRepo.Create(ctx, assistantMessage); err != nil {
		return nil, errors.NewDatabaseError("failed to save assistant message", err)
	}

	return &dto.SendMessageResponse{
		UserMessage: dto.MessageDTO{
			ID:        userMessage.ID,
			Role:      string(userMessage.Role),
			Content:   userMessage.Content,
			Metadata:  userMessage.Metadata,
			CreatedAt: userMessage.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		},
		AssistantMessage: dto.MessageDTO{
			ID:        assistantMessage.ID,
			Role:      string(assistantMessage.Role),
			Content:   assistantMessage.Content,
			Metadata:  assistantMessage.Metadata,
			CreatedAt: assistantMessage.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		},
	}, nil
}
