package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"customer-support-api/internal/domain/entities"
	"customer-support-api/internal/domain/repositories"
	"customer-support-api/internal/pkg/errors"
)

type returnRequestRepository struct {
	db *sql.DB
}

func NewReturnRequestRepository(db *sql.DB) repositories.ReturnRequestRepository {
	return &returnRequestRepository{db: db}
}

func (r *returnRequestRepository) Create(ctx context.Context, returnRequest *entities.ReturnRequest) error {
	query := `
		INSERT INTO return_requests (
			id, conversation_id, customer_id, order_number, items, reason, detailed_reason,
			photos, refund_method, delivery_method, collection_point, shipping_address,
			status, estimated_refund_date, refund_amount_cents, return_reference,
			created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	itemsJSON, err := json.Marshal(returnRequest.Items)
	if err != nil {
		return errors.NewDatabaseError("failed to marshal items", err)
	}

	photosJSON, err := json.Marshal(returnRequest.Photos)
	if err != nil {
		return errors.NewDatabaseError("failed to marshal photos", err)
	}

	// Handle nullable fields
	var conversationID, detailedReason, collectionPoint, shippingAddress sql.NullString
	var returnReference, estimatedRefundDate sql.NullString

	if returnRequest.ConversationID != nil {
		conversationID.Valid = true
		conversationID.String = *returnRequest.ConversationID
	}

	if returnRequest.DetailedReason != nil {
		detailedReason.Valid = true
		detailedReason.String = *returnRequest.DetailedReason
	}

	if returnRequest.CollectionPoint != nil {
		collectionPoint.Valid = true
		collectionPoint.String = *returnRequest.CollectionPoint
	}

	if returnRequest.ShippingAddress != nil {
		shippingAddress.Valid = true
		shippingAddress.String = *returnRequest.ShippingAddress
	}

	if returnRequest.ReturnReference != nil {
		returnReference.Valid = true
		returnReference.String = *returnRequest.ReturnReference
	}

	if returnRequest.EstimatedRefundDate != nil {
		estimatedRefundDate.Valid = true
		estimatedRefundDate.String = returnRequest.EstimatedRefundDate.Format(time.RFC3339)
	}

	_, err = r.db.ExecContext(ctx, query,
		returnRequest.ID,
		conversationID,
		returnRequest.CustomerID,
		returnRequest.OrderNumber,
		string(itemsJSON),
		returnRequest.Reason,
		detailedReason,
		string(photosJSON),
		returnRequest.RefundMethod,
		returnRequest.DeliveryMethod,
		collectionPoint,
		shippingAddress,
		returnRequest.Status,
		estimatedRefundDate,
		returnRequest.RefundAmountCents,
		returnReference,
		returnRequest.CreatedAt.Format(time.RFC3339),
		returnRequest.UpdatedAt.Format(time.RFC3339),
	)

	if err != nil {
		return errors.NewDatabaseError("failed to create return request", err)
	}

	return nil
}

func (r *returnRequestRepository) GetByID(ctx context.Context, id string) (*entities.ReturnRequest, error) {
	query := `
		SELECT id, conversation_id, customer_id, order_number, items, reason, detailed_reason,
			   photos, refund_method, delivery_method, collection_point, shipping_address,
			   status, estimated_refund_date, refund_amount_cents, return_reference,
			   created_at, updated_at
		FROM return_requests
		WHERE id = ?
	`

	var returnRequest entities.ReturnRequest
	var itemsJSON, photosJSON string
	var conversationID, detailedReason, collectionPoint, shippingAddress sql.NullString
	var returnReference, estimatedRefundDate sql.NullString
	var createdAt, updatedAt string

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&returnRequest.ID,
		&conversationID,
		&returnRequest.CustomerID,
		&returnRequest.OrderNumber,
		&itemsJSON,
		&returnRequest.Reason,
		&detailedReason,
		&photosJSON,
		&returnRequest.RefundMethod,
		&returnRequest.DeliveryMethod,
		&collectionPoint,
		&shippingAddress,
		&returnRequest.Status,
		&estimatedRefundDate,
		&returnRequest.RefundAmountCents,
		&returnReference,
		&createdAt,
		&updatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.NewNotFoundError("return request", id)
	}
	if err != nil {
		return nil, errors.NewDatabaseError("failed to get return request", err)
	}

	// Parse timestamps
	returnRequest.CreatedAt, err = time.Parse(time.RFC3339, createdAt)
	if err != nil {
		return nil, errors.NewDatabaseError("failed to parse created_at", err)
	}

	returnRequest.UpdatedAt, err = time.Parse(time.RFC3339, updatedAt)
	if err != nil {
		return nil, errors.NewDatabaseError("failed to parse updated_at", err)
	}

	// Parse optional string fields
	if conversationID.Valid {
		returnRequest.ConversationID = &conversationID.String
	}

	if detailedReason.Valid {
		returnRequest.DetailedReason = &detailedReason.String
	}

	if collectionPoint.Valid {
		returnRequest.CollectionPoint = &collectionPoint.String
	}

	if shippingAddress.Valid {
		returnRequest.ShippingAddress = &shippingAddress.String
	}

	if returnReference.Valid {
		returnRequest.ReturnReference = &returnReference.String
	}

	// Parse optional timestamp fields
	if estimatedRefundDate.Valid {
		t, err := time.Parse(time.RFC3339, estimatedRefundDate.String)
		if err != nil {
			return nil, errors.NewDatabaseError("failed to parse estimated_refund_date", err)
		}
		returnRequest.EstimatedRefundDate = &t
	}

	// Parse JSON fields
	if err := json.Unmarshal([]byte(itemsJSON), &returnRequest.Items); err != nil {
		return nil, errors.NewDatabaseError("failed to unmarshal items", err)
	}

	if err := json.Unmarshal([]byte(photosJSON), &returnRequest.Photos); err != nil {
		return nil, errors.NewDatabaseError("failed to unmarshal photos", err)
	}

	return &returnRequest, nil
}

func (r *returnRequestRepository) GetByCustomerID(ctx context.Context, customerID string) ([]*entities.ReturnRequest, error) {
	query := `
		SELECT id, conversation_id, customer_id, order_number, items, reason, detailed_reason,
			   photos, refund_method, delivery_method, collection_point, shipping_address,
			   status, estimated_refund_date, refund_amount_cents, return_reference,
			   created_at, updated_at
		FROM return_requests
		WHERE customer_id = ?
		ORDER BY created_at DESC
	`

	return r.queryReturnRequests(ctx, query, customerID)
}

func (r *returnRequestRepository) GetByOrderNumber(ctx context.Context, orderNumber string) ([]*entities.ReturnRequest, error) {
	query := `
		SELECT id, conversation_id, customer_id, order_number, items, reason, detailed_reason,
			   photos, refund_method, delivery_method, collection_point, shipping_address,
			   status, estimated_refund_date, refund_amount_cents, return_reference,
			   created_at, updated_at
		FROM return_requests
		WHERE order_number = ?
		ORDER BY created_at DESC
	`

	return r.queryReturnRequests(ctx, query, orderNumber)
}

func (r *returnRequestRepository) queryReturnRequests(ctx context.Context, query string, arg string) ([]*entities.ReturnRequest, error) {
	rows, err := r.db.QueryContext(ctx, query, arg)
	if err != nil {
		return nil, errors.NewDatabaseError("failed to query return requests", err)
	}
	defer rows.Close()

	var returnRequests []*entities.ReturnRequest

	for rows.Next() {
		var returnRequest entities.ReturnRequest
		var itemsJSON, photosJSON string
		var conversationID, detailedReason, collectionPoint, shippingAddress sql.NullString
		var returnReference, estimatedRefundDate sql.NullString
		var createdAt, updatedAt string

		err := rows.Scan(
			&returnRequest.ID,
			&conversationID,
			&returnRequest.CustomerID,
			&returnRequest.OrderNumber,
			&itemsJSON,
			&returnRequest.Reason,
			&detailedReason,
			&photosJSON,
			&returnRequest.RefundMethod,
			&returnRequest.DeliveryMethod,
			&collectionPoint,
			&shippingAddress,
			&returnRequest.Status,
			&estimatedRefundDate,
			&returnRequest.RefundAmountCents,
			&returnReference,
			&createdAt,
			&updatedAt,
		)
		if err != nil {
			return nil, errors.NewDatabaseError("failed to scan return request", err)
		}

		// Parse timestamps
		returnRequest.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		returnRequest.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)

		// Parse optional fields
		if conversationID.Valid {
			returnRequest.ConversationID = &conversationID.String
		}
		if detailedReason.Valid {
			returnRequest.DetailedReason = &detailedReason.String
		}
		if collectionPoint.Valid {
			returnRequest.CollectionPoint = &collectionPoint.String
		}
		if shippingAddress.Valid {
			returnRequest.ShippingAddress = &shippingAddress.String
		}
		if returnReference.Valid {
			returnRequest.ReturnReference = &returnReference.String
		}
		if estimatedRefundDate.Valid {
			t, _ := time.Parse(time.RFC3339, estimatedRefundDate.String)
			returnRequest.EstimatedRefundDate = &t
		}

		// Parse JSON fields
		json.Unmarshal([]byte(itemsJSON), &returnRequest.Items)
		json.Unmarshal([]byte(photosJSON), &returnRequest.Photos)

		returnRequests = append(returnRequests, &returnRequest)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.NewDatabaseError("error iterating return requests", err)
	}

	return returnRequests, nil
}

func (r *returnRequestRepository) Update(ctx context.Context, returnRequest *entities.ReturnRequest) error {
	query := `
		UPDATE return_requests
		SET status = ?, refund_amount_cents = ?, collection_point = ?, shipping_address = ?,
		    return_reference = ?, estimated_refund_date = ?, updated_at = ?
		WHERE id = ?
	`

	// Handle nullable fields
	var collectionPoint, shippingAddress, returnReference, estimatedRefundDate sql.NullString

	if returnRequest.CollectionPoint != nil {
		collectionPoint.Valid = true
		collectionPoint.String = *returnRequest.CollectionPoint
	}

	if returnRequest.ShippingAddress != nil {
		shippingAddress.Valid = true
		shippingAddress.String = *returnRequest.ShippingAddress
	}

	if returnRequest.ReturnReference != nil {
		returnReference.Valid = true
		returnReference.String = *returnRequest.ReturnReference
	}

	if returnRequest.EstimatedRefundDate != nil {
		estimatedRefundDate.Valid = true
		estimatedRefundDate.String = returnRequest.EstimatedRefundDate.Format(time.RFC3339)
	}

	returnRequest.UpdatedAt = time.Now()

	result, err := r.db.ExecContext(ctx, query,
		returnRequest.Status,
		returnRequest.RefundAmountCents,
		collectionPoint,
		shippingAddress,
		returnReference,
		estimatedRefundDate,
		returnRequest.UpdatedAt.Format(time.RFC3339),
		returnRequest.ID,
	)

	if err != nil {
		return errors.NewDatabaseError("failed to update return request", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.NewDatabaseError("failed to check rows affected", err)
	}

	if rowsAffected == 0 {
		return errors.NewNotFoundError("return request", returnRequest.ID)
	}

	return nil
}

func (r *returnRequestRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM return_requests WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return errors.NewDatabaseError("failed to delete return request", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.NewDatabaseError("failed to check rows affected", err)
	}

	if rowsAffected == 0 {
		return errors.NewNotFoundError("return request", id)
	}

	return nil
}
