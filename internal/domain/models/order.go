package models

import "time"

// OrderStatus represents the status of an order
type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "pending"
	OrderStatusConfirmed OrderStatus = "confirmed"
	OrderStatusShipped   OrderStatus = "shipped"
	OrderStatusDelivered OrderStatus = "delivered"
	OrderStatusCancelled OrderStatus = "cancelled"
)

// OrderItem represents an item in an order
type OrderItem struct {
	ID           string `json:"id"`
	ProductID    string `json:"product_id"`
	ProductName  string `json:"product_name"`
	SKU          string `json:"sku"`
	Quantity     int    `json:"quantity"`
	PriceCents   int    `json:"price_cents"`
	TotalCents   int    `json:"total_cents"`
	ImageURL     string `json:"image_url"`
	IsReturnable bool   `json:"is_returnable"`
}

// Order represents a customer order
type Order struct {
	ID              string      `json:"id"`
	OrderNumber     string      `json:"order_number"`
	CustomerID      string      `json:"customer_id"`
	Status          OrderStatus `json:"status"`
	Items           []OrderItem `json:"items"`
	SubtotalCents   int         `json:"subtotal_cents"`
	ShippingCents   int         `json:"shipping_cents"`
	TaxCents        int         `json:"tax_cents"`
	TotalCents      int         `json:"total_cents"`
	CurrencyCode    string      `json:"currency_code"`
	OrderDate       time.Time   `json:"order_date"`
	DeliveryDate    *time.Time  `json:"delivery_date,omitempty"`
	ShippingAddress Address     `json:"shipping_address"`
}

// Address represents a shipping or billing address
type Address struct {
	Line1      string `json:"line1"`
	Line2      string `json:"line2,omitempty"`
	City       string `json:"city"`
	State      string `json:"state"`
	PostalCode string `json:"postal_code"`
	Country    string `json:"country"`
}

// GetItemByID finds an order item by its ID
func (o *Order) GetItemByID(itemID string) *OrderItem {
	for i := range o.Items {
		if o.Items[i].ID == itemID {
			return &o.Items[i]
		}
	}
	return nil
}

// CalculateRefundAmount calculates the refund amount for given items
func (o *Order) CalculateRefundAmount(itemIDs []string) int {
	totalRefund := 0
	for _, itemID := range itemIDs {
		if item := o.GetItemByID(itemID); item != nil {
			totalRefund += item.TotalCents
		}
	}
	return totalRefund
}

// CanReturnItem checks if an item can be returned
func (o *Order) CanReturnItem(itemID string) bool {
	item := o.GetItemByID(itemID)
	if item == nil {
		return false
	}

	// Check if item is returnable
	if !item.IsReturnable {
		return false
	}

	// Check if order is in a returnable state
	if o.Status != OrderStatusDelivered {
		return false
	}

	// TODO: Add time-based validation (e.g., within 30 days of delivery)

	return true
}
