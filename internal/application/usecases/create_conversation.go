package usecases

import (
	"context"
	"customer-support-api/internal/application/dto"
	"customer-support-api/internal/domain/entities"
	"customer-support-api/internal/domain/repositories"
	"customer-support-api/internal/pkg/errors"
)

type CreateConversationUseCase struct {
	conversationRepo repositories.ConversationRepository
	messageRepo      repositories.MessageRepository
}

func NewCreateConversationUseCase(
	conversationRepo repositories.ConversationRepository,
	messageRepo repositories.MessageRepository,
) *CreateConversationUseCase {
	return &CreateConversationUseCase{
		conversationRepo: conversationRepo,
		messageRepo:      messageRepo,
	}
}

func (uc *CreateConversationUseCase) Execute(
	ctx context.Context,
	customerID string,
	req dto.CreateConversationRequest,
) (*dto.CreateConversationResponse, error) {
	var convType entities.ConversationType
	if req.Type == "natural_language" {
		convType = entities.NaturalLanguage
	} else {
		convType = entities.GuidedFlow
	}

	conversation := entities.NewConversation(customerID, convType)

	if err := uc.conversationRepo.Create(ctx, conversation); err != nil {
		return nil, errors.NewDatabaseError("failed to create conversation", err)
	}

	var messageDTO *dto.MessageDTO
	if req.InitialMessage != "" {
		userMsg := entities.NewMessage(conversation.ID, entities.RoleUser, req.InitialMessage)
		if err := uc.messageRepo.Create(ctx, userMsg); err != nil {
			return nil, errors.NewDatabaseError("failed to create message", err)
		}

		assistantMsg := entities.NewMessage(
			conversation.ID,
			entities.RoleAssistant,
			"Hello! I'm here to help you. How can I assist you today?",
		)
		if err := uc.messageRepo.Create(ctx, assistantMsg); err != nil {
			return nil, errors.NewDatabaseError("failed to create response message", err)
		}

		messageDTO = &dto.MessageDTO{
			Role:    string(assistantMsg.Role),
			Content: assistantMsg.Content,
		}
	}

	return &dto.CreateConversationResponse{
		ConversationID: conversation.ID,
		Type:           string(conversation.Type),
		Message:        messageDTO,
	}, nil
}
