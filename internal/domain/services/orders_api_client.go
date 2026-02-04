package services

import (
	"context"
	"customer-support-api/internal/domain/models"
)

// OrdersAPIClient defines the interface for interacting with the Orders API
type OrdersAPIClient interface {
	// GetOrderByNumber retrieves an order by its order number
	GetOrderByNumber(ctx context.Context, orderNumber string) (*models.Order, error)

	// GetOrderByID retrieves an order by its ID
	GetOrderByID(ctx context.Context, orderID string) (*models.Order, error)

	// ValidateOrderItems checks if the given items exist in the order and can be returned
	ValidateOrderItems(ctx context.Context, orderNumber string, itemIDs []string) error

	// CalculateRefundAmount calculates the refund amount for the given items
	CalculateRefundAmount(ctx context.Context, orderNumber string, itemIDs []string) (int, error)
}
