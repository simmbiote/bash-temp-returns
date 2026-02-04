package repositories

import (
	"context"
	"customer-support-api/internal/domain/entities"
)

type ReturnRequestRepository interface {
	Create(ctx context.Context, returnRequest *entities.ReturnRequest) error
	GetByID(ctx context.Context, id string) (*entities.ReturnRequest, error)
	GetByCustomerID(ctx context.Context, customerID string) ([]*entities.ReturnRequest, error)
	GetByOrderNumber(ctx context.Context, orderNumber string) ([]*entities.ReturnRequest, error)
	Update(ctx context.Context, returnRequest *entities.ReturnRequest) error
	Delete(ctx context.Context, id string) error
}
