package usecases

import (
	"context"
	"customer-support-api/internal/application/dto"
	"customer-support-api/internal/domain/repositories"
	"customer-support-api/internal/pkg/errors"
)

type GetConversationUseCase struct {
	convRepo    repositories.ConversationRepository
	messageRepo repositories.MessageRepository
}

func NewGetConversationUseCase(
	convRepo repositories.ConversationRepository,
	messageRepo repositories.MessageRepository,
) *GetConversationUseCase {
	return &GetConversationUseCase{
		convRepo:    convRepo,
		messageRepo: messageRepo,
	}
}

func (uc *GetConversationUseCase) Execute(
	ctx context.Context,
	conversationID string,
) (*dto.ConversationDetailsResponse, error) {
	// Get conversation
	conversation, err := uc.convRepo.GetByID(ctx, conversationID)
	if err != nil {
		return nil, err
	}

	// Get messages
	messages, err := uc.messageRepo.GetByConversationID(ctx, conversationID)
	if err != nil {
		return nil, errors.NewDatabaseError("failed to fetch messages", err)
	}

	// Convert to DTOs
	messageDTOs := make([]dto.MessageDTO, len(messages))
	for i, msg := range messages {
		messageDTOs[i] = dto.MessageDTO{
			ID:        msg.ID,
			Role:      string(msg.Role),
			Content:   msg.Content,
			Metadata:  msg.Metadata,
			CreatedAt: msg.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
	}

	var intent *string
	if conversation.Intent != nil {
		intentStr := string(*conversation.Intent)
		intent = &intentStr
	}

	var completedAt *string
	if conversation.CompletedAt != nil {
		completedAtStr := conversation.CompletedAt.Format("2006-01-02T15:04:05Z07:00")
		completedAt = &completedAtStr
	}

	return &dto.ConversationDetailsResponse{
		ID:          conversation.ID,
		CustomerID:  conversation.CustomerID,
		Type:        string(conversation.Type),
		Intent:      intent,
		CurrentStep: conversation.CurrentStep,
		FlowID:      conversation.FlowID,
		Context:     conversation.Context,
		Messages:    messageDTOs,
		CreatedAt:   conversation.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   conversation.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		CompletedAt: completedAt,
	}, nil
}
