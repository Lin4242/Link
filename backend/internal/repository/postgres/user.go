package postgres

import (
	"context"

	"link/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	query := `
		INSERT INTO users (password_hash, nickname, public_key, avatar_url)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at
	`
	return r.pool.QueryRow(ctx, query,
		user.PasswordHash,
		user.Nickname,
		user.PublicKey,
		user.AvatarURL,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
}

func (r *UserRepository) FindByID(ctx context.Context, id string) (*domain.User, error) {
	query := `
		SELECT id, password_hash, nickname, public_key, avatar_url, created_at, updated_at, last_seen_at
		FROM users WHERE id = $1
	`
	user := &domain.User{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.PasswordHash,
		&user.Nickname,
		&user.PublicKey,
		&user.AvatarURL,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.LastSeenAt,
	)
	if err == pgx.ErrNoRows {
		return nil, domain.ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserRepository) GetPublicKey(ctx context.Context, id string) (string, error) {
	var pk string
	err := r.pool.QueryRow(ctx, `SELECT public_key FROM users WHERE id = $1`, id).Scan(&pk)
	if err == pgx.ErrNoRows {
		return "", domain.ErrUserNotFound
	}
	return pk, err
}

func (r *UserRepository) Update(ctx context.Context, user *domain.User) error {
	query := `
		UPDATE users SET nickname = $2, avatar_url = $3, public_key = $4, updated_at = NOW()
		WHERE id = $1
	`
	_, err := r.pool.Exec(ctx, query, user.ID, user.Nickname, user.AvatarURL, user.PublicKey)
	return err
}

func (r *UserRepository) UpdateLastSeen(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `UPDATE users SET last_seen_at = NOW() WHERE id = $1`, id)
	return err
}

func (r *UserRepository) Search(ctx context.Context, query string, limit int) ([]*domain.User, error) {
	sql := `
		SELECT id, nickname, public_key, avatar_url, created_at, last_seen_at
		FROM users
		WHERE nickname ILIKE $1
		LIMIT $2
	`
	rows, err := r.pool.Query(ctx, sql, "%"+query+"%", limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*domain.User
	for rows.Next() {
		u := &domain.User{}
		if err := rows.Scan(&u.ID, &u.Nickname, &u.PublicKey, &u.AvatarURL, &u.CreatedAt, &u.LastSeenAt); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}
