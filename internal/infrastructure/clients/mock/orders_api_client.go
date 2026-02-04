package mock

import (
	"context"
	"fmt"
	"time"

	"customer-support-api/internal/domain/models"
	"customer-support-api/internal/domain/services"
	"customer-support-api/internal/pkg/errors"
)

type mockOrdersAPIClient struct {
	orders map[string]*models.Order
}

// NewMockOrdersAPIClient creates a mock Orders API client for testing/development
func NewMockOrdersAPIClient() services.OrdersAPIClient {
	client := &mockOrdersAPIClient{
		orders: make(map[string]*models.Order),
	}

	// Add some sample orders for testing
	client.seedSampleOrders()

	return client
}

func (c *mockOrdersAPIClient) seedSampleOrders() {
	now := time.Now()
	deliveryDate := now.Add(-5 * 24 * time.Hour) // Delivered 5 days ago

	// Sample order 1
	c.orders["ORD-123456"] = &models.Order{
		ID:          "order-123",
		OrderNumber: "ORD-123456",
		CustomerID:  "demo-customer-123",
		Status:      models.OrderStatusDelivered,
		Items: []models.OrderItem{
			{
				ID:           "ITEM-1",
				ProductID:    "PROD-456",
				ProductName:  "Nike Air Max 270",
				SKU:          "NIKE-AM270-BLK-42",
				Quantity:     1,
				PriceCents:   15999, // $159.99
				TotalCents:   15999,
				ImageURL:     "https://example.com/nike-air-max.jpg",
				IsReturnable: true,
			},
			{
				ID:           "ITEM-2",
				ProductID:    "PROD-789",
				ProductName:  "Adidas Running Socks (3-pack)",
				SKU:          "ADIDAS-SOCK-WHT",
				Quantity:     3,
				PriceCents:   1999, // $19.99
				TotalCents:   1999,
				ImageURL:     "https://example.com/adidas-socks.jpg",
				IsReturnable: true,
			},
		},
		SubtotalCents: 17998,
		ShippingCents: 500,
		TaxCents:      1620,
		TotalCents:    20118,
		CurrencyCode:  "USD",
		OrderDate:     now.Add(-10 * 24 * time.Hour),
		DeliveryDate:  &deliveryDate,
		ShippingAddress: models.Address{
			Line1:      "123 Main Street",
			City:       "Cape Town",
			State:      "Western Cape",
			PostalCode: "8001",
			Country:    "ZA",
		},
	}

	// Sample order 2
	c.orders["ORD-789012"] = &models.Order{
		ID:          "order-456",
		OrderNumber: "ORD-789012",
		CustomerID:  "demo-customer-123",
		Status:      models.OrderStatusDelivered,
		Items: []models.OrderItem{
			{
				ID:           "ITEM-3",
				ProductID:    "PROD-111",
				ProductName:  "Puma T-Shirt",
				SKU:          "PUMA-TSHIRT-BLU-L",
				Quantity:     2,
				PriceCents:   2999, // $29.99 each
				TotalCents:   5998,
				ImageURL:     "https://example.com/puma-tshirt.jpg",
				IsReturnable: true,
			},
		},
		SubtotalCents: 5998,
		ShippingCents: 500,
		TaxCents:      585,
		TotalCents:    7083,
		CurrencyCode:  "USD",
		OrderDate:     now.Add(-15 * 24 * time.Hour),
		DeliveryDate:  &deliveryDate,
		ShippingAddress: models.Address{
			Line1:      "456 Oak Avenue",
			City:       "Johannesburg",
			State:      "Gauteng",
			PostalCode: "2000",
			Country:    "ZA",
		},
	}

	// Sample order 3 - Not yet delivered
	c.orders["ORD-333444"] = &models.Order{
		ID:          "order-789",
		OrderNumber: "ORD-333444",
		CustomerID:  "demo-customer-123",
		Status:      models.OrderStatusShipped,
		Items: []models.OrderItem{
			{
				ID:           "ITEM-4",
				ProductID:    "PROD-222",
				ProductName:  "New Balance Sneakers",
				SKU:          "NB-574-GRY-44",
				Quantity:     1,
				PriceCents:   12999,
				TotalCents:   12999,
				ImageURL:     "https://example.com/nb-574.jpg",
				IsReturnable: true,
			},
		},
		SubtotalCents: 12999,
		ShippingCents: 500,
		TaxCents:      1215,
		TotalCents:    14714,
		CurrencyCode:  "USD",
		OrderDate:     now.Add(-2 * 24 * time.Hour),
		DeliveryDate:  nil,
		ShippingAddress: models.Address{
			Line1:      "789 Pine Road",
			City:       "Durban",
			State:      "KwaZulu-Natal",
			PostalCode: "4001",
			Country:    "ZA",
		},
	}
}

func (c *mockOrdersAPIClient) GetOrderByNumber(ctx context.Context, orderNumber string) (*models.Order, error) {
	order, exists := c.orders[orderNumber]
	if !exists {
		return nil, errors.NewNotFoundError("order", orderNumber)
	}
	return order, nil
}

func (c *mockOrdersAPIClient) GetOrderByID(ctx context.Context, orderID string) (*models.Order, error) {
	// Search through orders by ID
	for _, order := range c.orders {
		if order.ID == orderID {
			return order, nil
		}
	}
	return nil, errors.NewNotFoundError("order", orderID)
}

func (c *mockOrdersAPIClient) ValidateOrderItems(ctx context.Context, orderNumber string, itemIDs []string) error {
	order, err := c.GetOrderByNumber(ctx, orderNumber)
	if err != nil {
		return err
	}

	// Check each item exists and can be returned
	for _, itemID := range itemIDs {
		item := order.GetItemByID(itemID)
		if item == nil {
			return errors.NewInvalidInputError(
				fmt.Sprintf("item %s not found in order %s", itemID, orderNumber),
				nil,
			)
		}

		if !order.CanReturnItem(itemID) {
			return errors.NewInvalidInputError(
				fmt.Sprintf("item %s cannot be returned (order status: %s)", itemID, order.Status),
				nil,
			)
		}
	}

	return nil
}

func (c *mockOrdersAPIClient) CalculateRefundAmount(ctx context.Context, orderNumber string, itemIDs []string) (int, error) {
	order, err := c.GetOrderByNumber(ctx, orderNumber)
	if err != nil {
		return 0, err
	}

	return order.CalculateRefundAmount(itemIDs), nil
}
