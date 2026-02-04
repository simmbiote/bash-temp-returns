package dto

type CreateReturnRequest struct {
	ConversationID  *string         `json:"conversation_id,omitempty"`
	OrderNumber     string          `json:"order_number" binding:"required"`
	Items           []ReturnItemDTO `json:"items" binding:"required,min=1"`
	Reason          string          `json:"reason" binding:"required"`
	DetailedReason  *string         `json:"detailed_reason,omitempty"`
	Photos          []string        `json:"photos,omitempty"`
	RefundMethod    string          `json:"refund_method" binding:"required"`
	DeliveryMethod  string          `json:"delivery_method" binding:"required"`
	CollectionPoint *string         `json:"collection_point,omitempty"`
	ShippingAddress *string         `json:"shipping_address,omitempty"`
}

type ReturnItemDTO struct {
	OrderItemID string `json:"order_item_id" binding:"required"`
	ProductID   string `json:"product_id" binding:"required"`
	Quantity    int    `json:"quantity" binding:"required,min=1"`
}

type CreateReturnResponse struct {
	ID                  string  `json:"id"`
	ReturnReference     *string `json:"return_reference,omitempty"`
	Status              string  `json:"status"`
	EstimatedRefundDate *string `json:"estimated_refund_date,omitempty"`
	RefundAmountCents   int     `json:"refund_amount_cents"`
	Message             string  `json:"message"`
}

type ReturnRequestDTO struct {
	ID                  string          `json:"id"`
	ConversationID      string          `json:"conversation_id"`
	CustomerID          string          `json:"customer_id"`
	OrderNumber         string          `json:"order_number"`
	Items               []ReturnItemDTO `json:"items"`
	Reason              string          `json:"reason"`
	DetailedReason      *string         `json:"detailed_reason,omitempty"`
	Photos              []string        `json:"photos,omitempty"`
	RefundMethod        string          `json:"refund_method"`
	DeliveryMethod      string          `json:"delivery_method"`
	CollectionPoint     *string         `json:"collection_point,omitempty"`
	ShippingAddress     *string         `json:"shipping_address,omitempty"`
	Status              string          `json:"status"`
	EstimatedRefundDate *string         `json:"estimated_refund_date,omitempty"`
	RefundAmountCents   int             `json:"refund_amount_cents"`
	ReturnReference     *string         `json:"return_reference,omitempty"`
	CreatedAt           string          `json:"created_at"`
	UpdatedAt           string          `json:"updated_at"`
}
