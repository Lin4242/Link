package postgres

import (
	"context"
	"time"

	"link/internal/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

type MessageRepository struct {
	pool *pgxpool.Pool
}

func NewMessageRepository(pool *pgxpool.Pool) *MessageRepository {
	return &MessageRepository{pool: pool}
}

func (r *MessageRepository) Create(ctx context.Context, msg *domain.Message) error {
	query := `
		INSERT INTO messages (conversation_id, sender_id, encrypted_content)
		VALUES ($1, $2, $3)
		RETURNING id, created_at
	`
	return r.pool.QueryRow(ctx, query,
		msg.ConversationID, msg.SenderID, msg.EncryptedContent,
	).Scan(&msg.ID, &msg.CreatedAt)
}

func (r *MessageRepository) FindByConversation(ctx context.Context, convID string, limit int, before *time.Time) ([]*domain.Message, error) {
	var query string
	var args []interface{}

	if before != nil {
		query = `
			SELECT id, conversation_id, sender_id, encrypted_content, created_at, delivered_at, read_at
			FROM messages
			WHERE conversation_id = $1 AND created_at < $2
			ORDER BY created_at DESC
			LIMIT $3
		`
		args = []interface{}{convID, before, limit}
	} else {
		query = `
			SELECT id, conversation_id, sender_id, encrypted_content, created_at, delivered_at, read_at
			FROM messages
			WHERE conversation_id = $1
			ORDER BY created_at DESC
			LIMIT $2
		`
		args = []interface{}{convID, limit}
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []*domain.Message
	for rows.Next() {
		m := &domain.Message{}
		if err := rows.Scan(
			&m.ID, &m.ConversationID, &m.SenderID, &m.EncryptedContent,
			&m.CreatedAt, &m.DeliveredAt, &m.ReadAt,
		); err != nil {
			return nil, err
		}
		messages = append(messages, m)
	}

	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, rows.Err()
}

func (r *MessageRepository) MarkDelivered(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE messages SET delivered_at = NOW() WHERE id = $1 AND delivered_at IS NULL`,
		id,
	)
	return err
}

func (r *MessageRepository) MarkRead(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE messages SET read_at = NOW() WHERE id = $1 AND read_at IS NULL`,
		id,
	)
	return err
}

var _ domain.MessageRepository = (*MessageRepository)(nil)
