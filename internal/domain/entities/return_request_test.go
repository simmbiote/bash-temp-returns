package entities

import (
	"testing"
	"time"
)

func TestNewReturnRequest(t *testing.T) {
	customerID := "cust-123"
	orderNumber := "ORD-456"
	req := NewReturnRequest(customerID, orderNumber)

	if req.ID == "" {
		t.Error("Expected ID to be generated")
	}
	if req.CustomerID != customerID {
		t.Errorf("Expected CustomerID to be %s, got %s", customerID, req.CustomerID)
	}
	if req.OrderNumber != orderNumber {
		t.Errorf("Expected OrderNumber to be %s, got %s", orderNumber, req.OrderNumber)
	}
	if req.Status != ReturnStatusDraft {
		t.Errorf("Expected Status to be %s, got %s", ReturnStatusDraft, req.Status)
	}
	if req.Items == nil {
		t.Error("Expected Items to be initialized")
	}
	if req.RefundAmountCents != 0 {
		t.Errorf("Expected RefundAmountCents to be 0, got %d", req.RefundAmountCents)
	}
}

func TestReturnRequestAddItem(t *testing.T) {
	req := NewReturnRequest("cust-1", "ORD-1")

	err := req.AddItem("item-1", "prod-1", 2)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(req.Items) != 1 {
		t.Errorf("Expected 1 item, got %d", len(req.Items))
	}

	item := req.Items[0]
	if item.OrderItemID != "item-1" {
		t.Errorf("Expected OrderItemID to be item-1, got %s", item.OrderItemID)
	}
	if item.ProductID != "prod-1" {
		t.Errorf("Expected ProductID to be prod-1, got %s", item.ProductID)
	}
	if item.Quantity != 2 {
		t.Errorf("Expected Quantity to be 2, got %d", item.Quantity)
	}
}

func TestReturnRequestAddItemInvalidQuantity(t *testing.T) {
	req := NewReturnRequest("cust-1", "ORD-1")

	err := req.AddItem("item-1", "prod-1", 0)
	if err == nil {
		t.Error("Expected error for zero quantity")
	}

	err = req.AddItem("item-1", "prod-1", -1)
	if err == nil {
		t.Error("Expected error for negative quantity")
	}
}

func TestReturnRequestSetRefundAmount(t *testing.T) {
	req := NewReturnRequest("cust-1", "ORD-1")

	err := req.SetRefundAmount(5000)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if req.RefundAmountCents != 5000 {
		t.Errorf("Expected RefundAmountCents to be 5000, got %d", req.RefundAmountCents)
	}

	err = req.SetRefundAmount(-100)
	if err == nil {
		t.Error("Expected error for negative amount")
	}
}

func TestReturnRequestValidate(t *testing.T) {
	collectionPoint := "Store-123"
	shippingAddr := "123 Main St"

	tests := []struct {
		name    string
		req     *ReturnRequest
		wantErr bool
	}{
		{
			name: "valid collect return",
			req: &ReturnRequest{
				ID:                "ret-1",
				CustomerID:        "cust-1",
				OrderNumber:       "ORD-1",
				Items:             []ReturnItem{{OrderItemID: "item-1", ProductID: "prod-1", Quantity: 1}},
				Reason:            "defective",
				RefundMethod:      RefundMethodOriginal,
				DeliveryMethod:    DeliveryMethodCollect,
				CollectionPoint:   &collectionPoint,
				RefundAmountCents: 1000,
			},
			wantErr: false,
		},
		{
			name: "valid ship return",
			req: &ReturnRequest{
				ID:                "ret-1",
				CustomerID:        "cust-1",
				OrderNumber:       "ORD-1",
				Items:             []ReturnItem{{OrderItemID: "item-1", ProductID: "prod-1", Quantity: 1}},
				Reason:            "defective",
				RefundMethod:      RefundMethodOriginal,
				DeliveryMethod:    DeliveryMethodShip,
				ShippingAddress:   &shippingAddr,
				RefundAmountCents: 1000,
			},
			wantErr: false,
		},
		{
			name: "missing customer ID",
			req: &ReturnRequest{
				ID:          "ret-1",
				OrderNumber: "ORD-1",
				Items:       []ReturnItem{{OrderItemID: "item-1", ProductID: "prod-1", Quantity: 1}},
			},
			wantErr: true,
		},
		{
			name: "no items",
			req: &ReturnRequest{
				ID:          "ret-1",
				CustomerID:  "cust-1",
				OrderNumber: "ORD-1",
				Items:       []ReturnItem{},
			},
			wantErr: true,
		},
		{
			name: "item missing product ID",
			req: &ReturnRequest{
				ID:          "ret-1",
				CustomerID:  "cust-1",
				OrderNumber: "ORD-1",
				Items:       []ReturnItem{{OrderItemID: "item-1", Quantity: 1}},
			},
			wantErr: true,
		},
		{
			name: "collect without collection point",
			req: &ReturnRequest{
				ID:             "ret-1",
				CustomerID:     "cust-1",
				OrderNumber:    "ORD-1",
				Items:          []ReturnItem{{OrderItemID: "item-1", ProductID: "prod-1", Quantity: 1}},
				Reason:         "defective",
				RefundMethod:   RefundMethodOriginal,
				DeliveryMethod: DeliveryMethodCollect,
			},
			wantErr: true,
		},
		{
			name: "ship without shipping address",
			req: &ReturnRequest{
				ID:             "ret-1",
				CustomerID:     "cust-1",
				OrderNumber:    "ORD-1",
				Items:          []ReturnItem{{OrderItemID: "item-1", ProductID: "prod-1", Quantity: 1}},
				Reason:         "defective",
				RefundMethod:   RefundMethodOriginal,
				DeliveryMethod: DeliveryMethodShip,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestReturnRequestSubmit(t *testing.T) {
	req := NewReturnRequest("cust-1", "ORD-1")
	collectionPoint := "Store-123"

	// Should fail validation - no items
	err := req.Submit()
	if err == nil {
		t.Error("Expected error when submitting without items")
	}

	// Add required fields
	req.AddItem("item-1", "prod-1", 1)
	req.Reason = "defective"
	req.RefundMethod = RefundMethodOriginal
	req.DeliveryMethod = DeliveryMethodCollect
	req.CollectionPoint = &collectionPoint
	req.SetRefundAmount(1000)

	err = req.Submit()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if req.Status != ReturnStatusSubmitted {
		t.Errorf("Expected Status to be %s, got %s", ReturnStatusSubmitted, req.Status)
	}
}

func TestReturnRequestStatusTransitions(t *testing.T) {
	req := NewReturnRequest("cust-1", "ORD-1")

	// Approve
	refundDate := time.Now().Add(24 * 7 * time.Hour)
	req.Approve("RET-123", refundDate)
	if req.Status != ReturnStatusApproved {
		t.Errorf("Expected Status to be %s, got %s", ReturnStatusApproved, req.Status)
	}
	if req.ReturnReference == nil || *req.ReturnReference != "RET-123" {
		t.Error("Expected ReturnReference to be set")
	}

	// Mark collected
	req.MarkCollected()
	if req.Status != ReturnStatusCollected {
		t.Errorf("Expected Status to be %s, got %s", ReturnStatusCollected, req.Status)
	}

	// Mark received
	req.MarkReceived()
	if req.Status != ReturnStatusReceived {
		t.Errorf("Expected Status to be %s, got %s", ReturnStatusReceived, req.Status)
	}

	// Mark refunded
	req.MarkRefunded()
	if req.Status != ReturnStatusRefunded {
		t.Errorf("Expected Status to be %s, got %s", ReturnStatusRefunded, req.Status)
	}

	// Test reject
	req2 := NewReturnRequest("cust-2", "ORD-2")
	req2.Reject()
	if req2.Status != ReturnStatusRejected {
		t.Errorf("Expected Status to be %s, got %s", ReturnStatusRejected, req2.Status)
	}
}
