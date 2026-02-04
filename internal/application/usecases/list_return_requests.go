package usecases

import (
	"context"
	"customer-support-api/internal/application/dto"
	"customer-support-api/internal/domain/repositories"
)

type ListReturnRequestsUseCase struct {
	returnRepo repositories.ReturnRequestRepository
}

func NewListReturnRequestsUseCase(
	returnRepo repositories.ReturnRequestRepository,
) *ListReturnRequestsUseCase {
	return &ListReturnRequestsUseCase{
		returnRepo: returnRepo,
	}
}

func (uc *ListReturnRequestsUseCase) Execute(
	ctx context.Context,
	customerID string,
) ([]dto.ReturnRequestDTO, error) {
	returnReqs, err := uc.returnRepo.GetByCustomerID(ctx, customerID)
	if err != nil {
		return nil, err
	}

	// Convert to DTOs
	dtos := make([]dto.ReturnRequestDTO, len(returnReqs))
	for i, returnReq := range returnReqs {
		items := make([]dto.ReturnItemDTO, len(returnReq.Items))
		for j, item := range returnReq.Items {
			items[j] = dto.ReturnItemDTO{
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

		dtos[i] = dto.ReturnRequestDTO{
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
		}
	}

	return dtos, nil
}
