package usecases

import (
	"context"
	"customer-support-api/internal/application/dto"
	"customer-support-api/internal/domain/repositories"
)

type GetReturnRequestUseCase struct {
	returnRepo repositories.ReturnRequestRepository
}

func NewGetReturnRequestUseCase(
	returnRepo repositories.ReturnRequestRepository,
) *GetReturnRequestUseCase {
	return &GetReturnRequestUseCase{
		returnRepo: returnRepo,
	}
}

func (uc *GetReturnRequestUseCase) Execute(
	ctx context.Context,
	returnID string,
) (*dto.ReturnRequestDTO, error) {
	returnReq, err := uc.returnRepo.GetByID(ctx, returnID)
	if err != nil {
		return nil, err
	}

	// Convert to DTO
	items := make([]dto.ReturnItemDTO, len(returnReq.Items))
	for i, item := range returnReq.Items {
		items[i] = dto.ReturnItemDTO{
			OrderItemID: item.OrderItemID,
			ProductID:   item.ProductID,
			Quantity:    item.Quantity,
		}
	}

	var estimatedRefundDate *string
	if returnReq.EstimatedRefundDate != nil {
		dateStr := returnReq.EstimatedRefundDate.Format("2006-01-02T15:04:05Z07:00")
		estimatedRefundDate = &dateStr
	}

	conversationID := ""
	if returnReq.ConversationID != nil {
		conversationID = *returnReq.ConversationID
	}

	return &dto.ReturnRequestDTO{
		ID:                  returnReq.ID,
		ConversationID:      conversationID,
		CustomerID:          returnReq.CustomerID,
		OrderNumber:         returnReq.OrderNumber,
		Items:               items,
		Reason:              returnReq.Reason,
		DetailedReason:      returnReq.DetailedReason,
		Photos:              returnReq.Photos,
		RefundMethod:        string(returnReq.RefundMethod),
		DeliveryMethod:      string(returnReq.DeliveryMethod),
		CollectionPoint:     returnReq.CollectionPoint,
		ShippingAddress:     returnReq.ShippingAddress,
		Status:              string(returnReq.Status),
		EstimatedRefundDate: estimatedRefundDate,
		RefundAmountCents:   returnReq.RefundAmountCents,
		ReturnReference:     returnReq.ReturnReference,
		CreatedAt:           returnReq.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:           returnReq.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}
