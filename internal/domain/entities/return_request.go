package entities

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// ReturnStatus represents the status of a return request
type ReturnStatus string

const (
	ReturnStatusDraft     ReturnStatus = "draft"
	ReturnStatusSubmitted ReturnStatus = "submitted"
	ReturnStatusApproved  ReturnStatus = "approved"
	ReturnStatusRejected  ReturnStatus = "rejected"
	ReturnStatusCollected ReturnStatus = "collected"
	ReturnStatusReceived  ReturnStatus = "received"
	ReturnStatusRefunded  ReturnStatus = "refunded"
)

// RefundMethod represents the method of refund
type RefundMethod string

const (
	RefundMethodOriginal RefundMethod = "original_payment"
	RefundMethodGiftCard RefundMethod = "gift_card"
	RefundMethodBash     RefundMethod = "bash_account"
)

// DeliveryMethod represents the method of return delivery
type DeliveryMethod string

const (
	DeliveryMethodCollect DeliveryMethod = "collect"
	DeliveryMethodShip    DeliveryMethod = "ship"
)

// ReturnItem represents an item in a return request
type ReturnItem struct {
	OrderItemID string `json:"order_item_id"`
	ProductID   string `json:"product_id"`
	Quantity    int    `json:"quantity"`
}

// ReturnRequest represents a return request
type ReturnRequest struct {
	ID                  string         `json:"id"`
	ConversationID      *string        `json:"conversation_id,omitempty"`
	CustomerID          string         `json:"customer_id"`
	OrderNumber         string         `json:"order_number"`
	Items               []ReturnItem   `json:"items"`
	Reason              string         `json:"reason"`
	DetailedReason      *string        `json:"detailed_reason,omitempty"`
	Photos              []string       `json:"photos,omitempty"`
	RefundMethod        RefundMethod   `json:"refund_method"`
	DeliveryMethod      DeliveryMethod `json:"delivery_method"`
	CollectionPoint     *string        `json:"collection_point,omitempty"`
	ShippingAddress     *string        `json:"shipping_address,omitempty"`
	Status              ReturnStatus   `json:"status"`
	EstimatedRefundDate *time.Time     `json:"estimated_refund_date,omitempty"`
	RefundAmountCents   int            `json:"refund_amount_cents"`
	ReturnReference     *string        `json:"return_reference,omitempty"`
	CreatedAt           time.Time      `json:"created_at"`
	UpdatedAt           time.Time      `json:"updated_at"`
}

// NewReturnRequest creates a new return request
func NewReturnRequest(customerID, orderNumber string) *ReturnRequest {
	now := time.Now()
	return &ReturnRequest{
		ID:                uuid.New().String(),
		CustomerID:        customerID,
		OrderNumber:       orderNumber,
		Items:             []ReturnItem{},
		Photos:            []string{},
		Status:            ReturnStatusDraft,
		RefundAmountCents: 0,
		CreatedAt:         now,
		UpdatedAt:         now,
	}
}

// Validate validates the return request
func (r *ReturnRequest) Validate() error {
	if r.ID == "" {
		return errors.New("return request ID is required")
	}
	if r.CustomerID == "" {
		return errors.New("customer ID is required")
	}
	if r.OrderNumber == "" {
		return errors.New("order number is required")
	}
	if len(r.Items) == 0 {
		return errors.New("at least one item is required")
	}
	for _, item := range r.Items {
		if item.OrderItemID == "" {
			return errors.New("order item ID is required for all items")
		}
		if item.ProductID == "" {
			return errors.New("product ID is required for all items")
		}
		if item.Quantity <= 0 {
			return errors.New("quantity must be greater than 0")
		}
	}
	if r.Reason == "" {
		return errors.New("reason is required")
	}
	if r.RefundMethod == "" {
		return errors.New("refund method is required")
	}
	if r.DeliveryMethod == "" {
		return errors.New("delivery method is required")
	}
	if r.DeliveryMethod == DeliveryMethodCollect && r.CollectionPoint == nil {
		return errors.New("collection point is required for collect delivery method")
	}
	if r.DeliveryMethod == DeliveryMethodShip && r.ShippingAddress == nil {
		return errors.New("shipping address is required for ship delivery method")
	}
	return nil
}

// AddItem adds an item to the return request
func (r *ReturnRequest) AddItem(orderItemID, productID string, quantity int) error {
	if quantity <= 0 {
		return errors.New("quantity must be greater than 0")
	}
	r.Items = append(r.Items, ReturnItem{
		OrderItemID: orderItemID,
		ProductID:   productID,
		Quantity:    quantity,
	})
	r.UpdatedAt = time.Now()
	return nil
}

// AddPhoto adds a photo to the return request
func (r *ReturnRequest) AddPhoto(photoURL string) {
	r.Photos = append(r.Photos, photoURL)
	r.UpdatedAt = time.Now()
}

// SetRefundAmount sets the refund amount
func (r *ReturnRequest) SetRefundAmount(amountCents int) error {
	if amountCents < 0 {
		return errors.New("refund amount cannot be negative")
	}
	r.RefundAmountCents = amountCents
	r.UpdatedAt = time.Now()
	return nil
}

// Submit submits the return request
func (r *ReturnRequest) Submit() error {
	if err := r.Validate(); err != nil {
		return err
	}
	r.Status = ReturnStatusSubmitted
	r.UpdatedAt = time.Now()
	return nil
}

// Approve approves the return request
func (r *ReturnRequest) Approve(returnReference string, estimatedRefundDate time.Time) {
	r.Status = ReturnStatusApproved
	r.ReturnReference = &returnReference
	r.EstimatedRefundDate = &estimatedRefundDate
	r.UpdatedAt = time.Now()
}

// Reject rejects the return request
func (r *ReturnRequest) Reject() {
	r.Status = ReturnStatusRejected
	r.UpdatedAt = time.Now()
}

// MarkCollected marks the return as collected
func (r *ReturnRequest) MarkCollected() {
	r.Status = ReturnStatusCollected
	r.UpdatedAt = time.Now()
}

// MarkReceived marks the return as received
func (r *ReturnRequest) MarkReceived() {
	r.Status = ReturnStatusReceived
	r.UpdatedAt = time.Now()
}

// MarkRefunded marks the return as refunded
func (r *ReturnRequest) MarkRefunded() {
	r.Status = ReturnStatusRefunded
	r.UpdatedAt = time.Now()
}
