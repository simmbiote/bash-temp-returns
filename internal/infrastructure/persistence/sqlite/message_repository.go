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

type messageRepository struct {
	db *sql.DB
}

func NewMessageRepository(db *sql.DB) repositories.MessageRepository {
	return &messageRepository{db: db}
}

func (r *messageRepository) Create(ctx context.Context, message *entities.Message) error {
	query := `
		INSERT INTO messages (
			id, conversation_id, role, content, metadata, created_at
		) VALUES (?, ?, ?, ?, ?, ?)
	`

	metadataJSON, err := json.Marshal(message.Metadata)
	if err != nil {
		return errors.NewDatabaseError("failed to marshal metadata", err)
	}

	_, err = r.db.ExecContext(ctx, query,
		message.ID,
		message.ConversationID,
		message.Role,
		message.Content,
		string(metadataJSON),
		message.CreatedAt.Format(time.RFC3339),
	)

	if err != nil {
		return errors.NewDatabaseError("failed to create message", err)
	}

	return nil
}

func (r *messageRepository) GetByID(ctx context.Context, id string) (*entities.Message, error) {
	query := `
		SELECT id, conversation_id, role, content, metadata, created_at
		FROM messages
		WHERE id = ?
	`

	var message entities.Message
	var metadataJSON string
	var createdAt string

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&message.ID,
		&message.ConversationID,
		&message.Role,
		&message.Content,
		&metadataJSON,
		&createdAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.NewNotFoundError("message", id)
	}
	if err != nil {
		return nil, errors.NewDatabaseError("failed to get message", err)
	}

	// Parse timestamp
	message.CreatedAt, err = time.Parse(time.RFC3339, createdAt)
	if err != nil {
		return nil, errors.NewDatabaseError("failed to parse created_at", err)
	}

	// Parse metadata JSON
	if err := json.Unmarshal([]byte(metadataJSON), &message.Metadata); err != nil {
		return nil, errors.NewDatabaseError("failed to unmarshal metadata", err)
	}

	return &message, nil
}

func (r *messageRepository) GetByConversationID(ctx context.Context, conversationID string) ([]*entities.Message, error) {
	query := `
		SELECT id, conversation_id, role, content, metadata, created_at
		FROM messages
		WHERE conversation_id = ?
		ORDER BY created_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query, conversationID)
	if err != nil {
		return nil, errors.NewDatabaseError("failed to query messages", err)
	}
	defer rows.Close()

	var messages []*entities.Message

	for rows.Next() {
		var message entities.Message
		var metadataJSON string
		var createdAt string

		err := rows.Scan(
			&message.ID,
			&message.ConversationID,
			&message.Role,
			&message.Content,
			&metadataJSON,
			&createdAt,
		)
		if err != nil {
			return nil, errors.NewDatabaseError("failed to scan message", err)
		}

		// Parse timestamp
		message.CreatedAt, err = time.Parse(time.RFC3339, createdAt)
		if err != nil {
			return nil, errors.NewDatabaseError("failed to parse created_at", err)
		}

		// Parse metadata JSON
		if err := json.Unmarshal([]byte(metadataJSON), &message.Metadata); err != nil {
			return nil, errors.NewDatabaseError("failed to unmarshal metadata", err)
		}

		messages = append(messages, &message)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.NewDatabaseError("error iterating messages", err)
	}

	return messages, nil
}

func (r *messageRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM messages WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return errors.NewDatabaseError("failed to delete message", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.NewDatabaseError("failed to check rows affected", err)
	}

	if rowsAffected == 0 {
		return errors.NewNotFoundError("message", id)
	}

	return nil
}
