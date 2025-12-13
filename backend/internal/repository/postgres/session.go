package postgres

import (
	"context"

	"link/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SessionRepository struct {
	pool *pgxpool.Pool
}

func NewSessionRepository(pool *pgxpool.Pool) *SessionRepository {
	return &SessionRepository{pool: pool}
}

func (r *SessionRepository) Create(ctx context.Context, session *domain.Session) error {
	query := `
		INSERT INTO sessions (user_id, token_hash, expires_at)
		VALUES ($1, $2, $3)
		RETURNING id, created_at
	`
	return r.pool.QueryRow(ctx, query,
		session.UserID, session.TokenHash, session.ExpiresAt,
	).Scan(&session.ID, &session.CreatedAt)
}

func (r *SessionRepository) FindByTokenHash(ctx context.Context, hash string) (*domain.Session, error) {
	query := `
		SELECT id, user_id, token_hash, created_at, expires_at, revoked_at
		FROM sessions WHERE token_hash = $1
	`
	s := &domain.Session{}
	err := r.pool.QueryRow(ctx, query, hash).Scan(
		&s.ID, &s.UserID, &s.TokenHash, &s.CreatedAt, &s.ExpiresAt, &s.RevokedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return s, err
}

func (r *SessionRepository) RevokeAllByUser(ctx context.Context, userID string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE sessions SET revoked_at = NOW() WHERE user_id = $1 AND revoked_at IS NULL`,
		userID,
	)
	return err
}

func (r *SessionRepository) Revoke(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `UPDATE sessions SET revoked_at = NOW() WHERE id = $1`, id)
	return err
}

func (r *SessionRepository) CleanupExpired(ctx context.Context) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM sessions WHERE expires_at < NOW()`)
	return err
}

var _ domain.SessionRepository = (*SessionRepository)(nil)
