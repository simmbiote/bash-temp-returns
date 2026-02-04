package repositories

import (
	"context"
	"customer-support-api/internal/domain/entities"
)

type MessageRepository interface {
	Create(ctx context.Context, message *entities.Message) error
	GetByID(ctx context.Context, id string) (*entities.Message, error)
	GetByConversationID(ctx context.Context, conversationID string) ([]*entities.Message, error)
	Delete(ctx context.Context, id string) error
}
