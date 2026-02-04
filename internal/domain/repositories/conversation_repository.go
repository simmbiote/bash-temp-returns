package repositories

import (
	"context"
	"customer-support-api/internal/domain/entities"
)

type ConversationRepository interface {
	Create(ctx context.Context, conversation *entities.Conversation) error
	GetByID(ctx context.Context, id string) (*entities.Conversation, error)
	GetByCustomerID(ctx context.Context, customerID string) ([]*entities.Conversation, error)
	Update(ctx context.Context, conversation *entities.Conversation) error
	Delete(ctx context.Context, id string) error
}
