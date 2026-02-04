package usecases

import (
	"context"
	"customer-support-api/internal/application/dto"
	"customer-support-api/internal/domain/entities"
	"customer-support-api/internal/domain/repositories"
	"customer-support-api/internal/domain/services"
	"customer-support-api/internal/pkg/errors"
	"time"
)

type CreateReturnRequestUseCase struct {
	returnRepo      repositories.ReturnRequestRepository
	convRepo        repositories.ConversationRepository
	ordersAPIClient services.OrdersAPIClient
}

func NewCreateReturnRequestUseCase(
	returnRepo repositories.ReturnRequestRepository,
	convRepo repositories.ConversationRepository,
	ordersAPIClient services.OrdersAPIClient,
) *CreateReturnRequestUseCase {
	return &CreateReturnRequestUseCase{
		returnRepo:      returnRepo,
		convRepo:        convRepo,
		ordersAPIClient: ordersAPIClient,
	}
}

func (uc *CreateReturnRequestUseCase) Execute(
	ctx context.Context,
	customerID string,
	req dto.CreateReturnRequest,
) (*dto.CreateReturnResponse, error) {
	// Step 1: Validate order exists and fetch order details
	order, err := uc.ordersAPIClient.GetOrderByNumber(ctx, req.OrderNumber)
	if err != nil {
		return nil, errors.NewValidationError("order not found", err)
	}

	// Step 2: Verify customer owns this order
	if order.CustomerID != customerID {
		return nil, errors.NewInvalidInputError("order does not belong to customer", nil)
	}

	// Step 3: Extract item IDs from request
	itemIDs := make([]string, len(req.Items))
	for i, item := range req.Items {
		itemIDs[i] = item.OrderItemID
	}

	// Step 4: Validate items can be returned
	if err := uc.ordersAPIClient.ValidateOrderItems(ctx, req.OrderNumber, itemIDs); err != nil {
		return nil, err
	}

	// Step 5: Calculate refund amount
	refundAmount, err := uc.ordersAPIClient.CalculateRefundAmount(ctx, req.OrderNumber, itemIDs)
	if err != nil {
		return nil, errors.NewDatabaseError("failed to calculate refund", err)
	}

	// Step 6: Create return request entity
	returnReq := entities.NewReturnRequest(customerID, req.OrderNumber)

	if req.ConversationID != nil {
		returnReq.ConversationID = req.ConversationID
	}

	for _, item := range req.Items {
		if err := returnReq.AddItem(item.OrderItemID, item.ProductID, item.Quantity); err != nil {
			return nil, errors.NewValidationError("invalid item", err)
		}
	}

	returnReq.Reason = req.Reason
	if req.DetailedReason != nil {
		returnReq.DetailedReason = req.DetailedReason
	}
	if req.Photos != nil {
		returnReq.Photos = req.Photos
	}

	returnReq.RefundMethod = entities.RefundMethod(req.RefundMethod)
	returnReq.DeliveryMethod = entities.DeliveryMethod(req.DeliveryMethod)

	if req.CollectionPoint != nil {
		returnReq.CollectionPoint = req.CollectionPoint
	}
	if req.ShippingAddress != nil {
		returnReq.ShippingAddress = req.ShippingAddress
	}

	if err := returnReq.Validate(); err != nil {
		return nil, errors.NewValidationError("validation failed", err)
	}

	if err := returnReq.Submit(); err != nil {
		return nil, errors.NewValidationError("failed to submit return", err)
	}

	// Set the calculated refund amount
	returnReq.SetRefundAmount(refundAmount)

	estimatedDate := time.Now().Add(7 * 24 * time.Hour)
	returnReference := "RET-" + returnReq.ID[:8]
	returnReq.Approve(returnReference, estimatedDate)

	if err := uc.returnRepo.Create(ctx, returnReq); err != nil {
		return nil, errors.NewDatabaseError("failed to create return request", err)
	}

	estimatedDateStr := estimatedDate.Format(time.RFC3339)
	return &dto.CreateReturnResponse{
		ID:                  returnReq.ID,
		ReturnReference:     returnReq.ReturnReference,
		Status:              string(returnReq.Status),
		EstimatedRefundDate: &estimatedDateStr,
		RefundAmountCents:   returnReq.RefundAmountCents,
		Message:             "Your return has been submitted successfully.",
	}, nil
}
