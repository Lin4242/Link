package postgres

import (
	"context"

	"link/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ConversationRepository struct {
	pool *pgxpool.Pool
}

func NewConversationRepository(pool *pgxpool.Pool) *ConversationRepository {
	return &ConversationRepository{pool: pool}
}

func (r *ConversationRepository) Create(ctx context.Context, c *domain.Conversation) error {
	p1, p2 := c.Participant1, c.Participant2
	if p1 > p2 {
		p1, p2 = p2, p1
	}
	query := `
		INSERT INTO conversations (participant_1, participant_2)
		VALUES ($1, $2)
		RETURNING id, created_at
	`
	return r.pool.QueryRow(ctx, query, p1, p2).Scan(&c.ID, &c.CreatedAt)
}

func (r *ConversationRepository) FindByID(ctx context.Context, id string) (*domain.Conversation, error) {
	query := `
		SELECT id, participant_1, participant_2, last_message_at, created_at
		FROM conversations WHERE id = $1
	`
	c := &domain.Conversation{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&c.ID, &c.Participant1, &c.Participant2, &c.LastMessageAt, &c.CreatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, domain.ErrConversationNotFound
	}
	return c, err
}

func (r *ConversationRepository) FindByParticipants(ctx context.Context, userA, userB string) (*domain.Conversation, error) {
	p1, p2 := userA, userB
	if p1 > p2 {
		p1, p2 = p2, p1
	}
	query := `
		SELECT id, participant_1, participant_2, last_message_at, created_at
		FROM conversations WHERE participant_1 = $1 AND participant_2 = $2
	`
	c := &domain.Conversation{}
	err := r.pool.QueryRow(ctx, query, p1, p2).Scan(
		&c.ID, &c.Participant1, &c.Participant2, &c.LastMessageAt, &c.CreatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return c, err
}

func (r *ConversationRepository) FindByUser(ctx context.Context, userID string) ([]*domain.ConversationWithPeer, error) {
	query := `
		SELECT c.id, c.participant_1, c.participant_2, c.last_message_at, c.created_at,
		       u.id, u.nickname, u.public_key, u.avatar_url, u.last_seen_at,
		       COALESCE((
		           SELECT COUNT(*) FROM messages m
		           WHERE m.conversation_id = c.id AND m.sender_id != $1 AND m.read_at IS NULL
		       ), 0) as unread
		FROM conversations c
		JOIN users u ON (
			CASE
				WHEN c.participant_1 = $1 THEN c.participant_2 = u.id
				ELSE c.participant_1 = u.id
			END
		)
		WHERE c.participant_1 = $1 OR c.participant_2 = $1
		ORDER BY c.last_message_at DESC NULLS LAST
	`
	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*domain.ConversationWithPeer
	for rows.Next() {
		cw := &domain.ConversationWithPeer{Peer: &domain.User{}}
		err := rows.Scan(
			&cw.ID, &cw.Participant1, &cw.Participant2, &cw.LastMessageAt, &cw.CreatedAt,
			&cw.Peer.ID, &cw.Peer.Nickname, &cw.Peer.PublicKey, &cw.Peer.AvatarURL, &cw.Peer.LastSeenAt,
			&cw.UnreadCount,
		)
		if err != nil {
			return nil, err
		}
		result = append(result, cw)
	}
	return result, rows.Err()
}

func (r *ConversationRepository) GetOrCreate(ctx context.Context, userA, userB string) (*domain.Conversation, error) {
	conv, err := r.FindByParticipants(ctx, userA, userB)
	if err != nil {
		return nil, err
	}
	if conv != nil {
		return conv, nil
	}

	conv = &domain.Conversation{
		Participant1: userA,
		Participant2: userB,
	}
	if err := r.Create(ctx, conv); err != nil {
		return r.FindByParticipants(ctx, userA, userB)
	}
	return conv, nil
}

var _ domain.ConversationRepository = (*ConversationRepository)(nil)
