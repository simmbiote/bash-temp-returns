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

type conversationRepository struct {
	db *sql.DB
}

func NewConversationRepository(db *sql.DB) repositories.ConversationRepository {
	return &conversationRepository{db: db}
}

func (r *conversationRepository) Create(ctx context.Context, conversation *entities.Conversation) error {
	query := `
		INSERT INTO conversations (
			id, customer_id, type, intent, current_step, flow_id, context,
			created_at, updated_at, completed_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	contextJSON, err := json.Marshal(conversation.Context)
	if err != nil {
		return errors.NewDatabaseError("failed to marshal context", err)
	}

	var intent, currentStep, flowID, completedAt sql.NullString

	if conversation.Intent != nil {
		intent.Valid = true
		intent.String = string(*conversation.Intent)
	}

	if conversation.CurrentStep != nil {
		currentStep.Valid = true
		currentStep.String = *conversation.CurrentStep
	}

	if conversation.FlowID != nil {
		flowID.Valid = true
		flowID.String = *conversation.FlowID
	}

	if conversation.CompletedAt != nil {
		completedAt.Valid = true
		completedAt.String = conversation.CompletedAt.Format(time.RFC3339)
	}

	_, err = r.db.ExecContext(ctx, query,
		conversation.ID,
		conversation.CustomerID,
		conversation.Type,
		intent,
		currentStep,
		flowID,
		string(contextJSON),
		conversation.CreatedAt.Format(time.RFC3339),
		conversation.UpdatedAt.Format(time.RFC3339),
		completedAt,
	)

	if err != nil {
		return errors.NewDatabaseError("failed to create conversation", err)
	}

	return nil
}

func (r *conversationRepository) GetByID(ctx context.Context, id string) (*entities.Conversation, error) {
	query := `
		SELECT id, customer_id, type, intent, current_step, flow_id, context,
			   created_at, updated_at, completed_at
		FROM conversations
		WHERE id = ?
	`

	var conversation entities.Conversation
	var contextJSON string
	var intent, currentStep, flowID, completedAt sql.NullString
	var createdAt, updatedAt string

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&conversation.ID,
		&conversation.CustomerID,
		&conversation.Type,
		&intent,
		&currentStep,
		&flowID,
		&contextJSON,
		&createdAt,
		&updatedAt,
		&completedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.NewNotFoundError("conversation", id)
	}
	if err != nil {
		return nil, errors.NewDatabaseError("failed to get conversation", err)
	}

	// Parse timestamps
	conversation.CreatedAt, err = time.Parse(time.RFC3339, createdAt)
	if err != nil {
		return nil, errors.NewDatabaseError("failed to parse created_at", err)
	}

	conversation.UpdatedAt, err = time.Parse(time.RFC3339, updatedAt)
	if err != nil {
		return nil, errors.NewDatabaseError("failed to parse updated_at", err)
	}

	// Parse optional fields
	if intent.Valid {
		intentValue := entities.Intent(intent.String)
		conversation.Intent = &intentValue
	}

	if currentStep.Valid {
		conversation.CurrentStep = &currentStep.String
	}

	if flowID.Valid {
		conversation.FlowID = &flowID.String
	}

	if completedAt.Valid {
		t, err := time.Parse(time.RFC3339, completedAt.String)
		if err != nil {
			return nil, errors.NewDatabaseError("failed to parse completed_at", err)
		}
		conversation.CompletedAt = &t
	}

	// Parse context JSON
	if err := json.Unmarshal([]byte(contextJSON), &conversation.Context); err != nil {
		return nil, errors.NewDatabaseError("failed to unmarshal context", err)
	}

	return &conversation, nil
}

func (r *conversationRepository) GetByCustomerID(ctx context.Context, customerID string) ([]*entities.Conversation, error) {
	query := `
		SELECT id, customer_id, type, intent, current_step, flow_id, context,
			   created_at, updated_at, completed_at
		FROM conversations
		WHERE customer_id = ?
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, customerID)
	if err != nil {
		return nil, errors.NewDatabaseError("failed to query conversations", err)
	}
	defer rows.Close()

	var conversations []*entities.Conversation

	for rows.Next() {
		var conversation entities.Conversation
		var contextJSON string
		var intent, currentStep, flowID, completedAt sql.NullString
		var createdAt, updatedAt string

		err := rows.Scan(
			&conversation.ID,
			&conversation.CustomerID,
			&conversation.Type,
			&intent,
			&currentStep,
			&flowID,
			&contextJSON,
			&createdAt,
			&updatedAt,
			&completedAt,
		)
		if err != nil {
			return nil, errors.NewDatabaseError("failed to scan conversation", err)
		}

		// Parse timestamps
		conversation.CreatedAt, err = time.Parse(time.RFC3339, createdAt)
		if err != nil {
			return nil, errors.NewDatabaseError("failed to parse created_at", err)
		}

		conversation.UpdatedAt, err = time.Parse(time.RFC3339, updatedAt)
		if err != nil {
			return nil, errors.NewDatabaseError("failed to parse updated_at", err)
		}

		// Parse optional fields
		if intent.Valid {
			intentValue := entities.Intent(intent.String)
			conversation.Intent = &intentValue
		}

		if currentStep.Valid {
			conversation.CurrentStep = &currentStep.String
		}

		if flowID.Valid {
			conversation.FlowID = &flowID.String
		}

		if completedAt.Valid {
			t, err := time.Parse(time.RFC3339, completedAt.String)
			if err != nil {
				return nil, errors.NewDatabaseError("failed to parse completed_at", err)
			}
			conversation.CompletedAt = &t
		}

		// Parse context JSON
		if err := json.Unmarshal([]byte(contextJSON), &conversation.Context); err != nil {
			return nil, errors.NewDatabaseError("failed to unmarshal context", err)
		}

		conversations = append(conversations, &conversation)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.NewDatabaseError("error iterating conversations", err)
	}

	return conversations, nil
}

func (r *conversationRepository) Update(ctx context.Context, conversation *entities.Conversation) error {
	query := `
		UPDATE conversations
		SET intent = ?, current_step = ?, flow_id = ?, context = ?, completed_at = ?, updated_at = ?
		WHERE id = ?
	`

	contextJSON, err := json.Marshal(conversation.Context)
	if err != nil {
		return errors.NewDatabaseError("failed to marshal context", err)
	}

	var intent, currentStep, flowID, completedAt sql.NullString

	if conversation.Intent != nil {
		intent.Valid = true
		intent.String = string(*conversation.Intent)
	}

	if conversation.CurrentStep != nil {
		currentStep.Valid = true
		currentStep.String = *conversation.CurrentStep
	}

	if conversation.FlowID != nil {
		flowID.Valid = true
		flowID.String = *conversation.FlowID
	}

	if conversation.CompletedAt != nil {
		completedAt.Valid = true
		completedAt.String = conversation.CompletedAt.Format(time.RFC3339)
	}

	conversation.UpdatedAt = time.Now()

	result, err := r.db.ExecContext(ctx, query,
		intent,
		currentStep,
		flowID,
		string(contextJSON),
		completedAt,
		conversation.UpdatedAt.Format(time.RFC3339),
		conversation.ID,
	)

	if err != nil {
		return errors.NewDatabaseError("failed to update conversation", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.NewDatabaseError("failed to check rows affected", err)
	}

	if rowsAffected == 0 {
		return errors.NewNotFoundError("conversation", conversation.ID)
	}

	return nil
}

func (r *conversationRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM conversations WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return errors.NewDatabaseError("failed to delete conversation", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.NewDatabaseError("failed to check rows affected", err)
	}

	if rowsAffected == 0 {
		return errors.NewNotFoundError("conversation", id)
	}

	return nil
}
