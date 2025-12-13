package postgres

import (
	"context"

	"link/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type FriendshipRepository struct {
	pool *pgxpool.Pool
}

func NewFriendshipRepository(pool *pgxpool.Pool) *FriendshipRepository {
	return &FriendshipRepository{pool: pool}
}

func (r *FriendshipRepository) Create(ctx context.Context, f *domain.Friendship) error {
	query := `
		INSERT INTO friendships (requester_id, addressee_id, status)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at
	`
	return r.pool.QueryRow(ctx, query,
		f.RequesterID, f.AddresseeID, f.Status,
	).Scan(&f.ID, &f.CreatedAt, &f.UpdatedAt)
}

func (r *FriendshipRepository) FindByUsers(ctx context.Context, userA, userB string) (*domain.Friendship, error) {
	query := `
		SELECT id, requester_id, addressee_id, status, created_at, updated_at
		FROM friendships
		WHERE (requester_id = $1 AND addressee_id = $2)
		   OR (requester_id = $2 AND addressee_id = $1)
	`
	f := &domain.Friendship{}
	err := r.pool.QueryRow(ctx, query, userA, userB).Scan(
		&f.ID, &f.RequesterID, &f.AddresseeID, &f.Status, &f.CreatedAt, &f.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return f, err
}

func (r *FriendshipRepository) FindFriends(ctx context.Context, userID string) ([]*domain.FriendWithUser, error) {
	query := `
		SELECT f.id, f.requester_id, f.addressee_id, f.status, f.created_at, f.updated_at,
		       u.id, u.nickname, u.public_key, u.avatar_url, u.last_seen_at
		FROM friendships f
		JOIN users u ON (
			CASE
				WHEN f.requester_id = $1 THEN f.addressee_id = u.id
				ELSE f.requester_id = u.id
			END
		)
		WHERE (f.requester_id = $1 OR f.addressee_id = $1)
		  AND f.status = 'accepted'
	`
	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*domain.FriendWithUser
	for rows.Next() {
		fw := &domain.FriendWithUser{Friend: &domain.User{}}
		err := rows.Scan(
			&fw.ID, &fw.RequesterID, &fw.AddresseeID, &fw.Status, &fw.CreatedAt, &fw.UpdatedAt,
			&fw.Friend.ID, &fw.Friend.Nickname, &fw.Friend.PublicKey, &fw.Friend.AvatarURL, &fw.Friend.LastSeenAt,
		)
		if err != nil {
			return nil, err
		}
		result = append(result, fw)
	}
	return result, rows.Err()
}

func (r *FriendshipRepository) FindPendingRequests(ctx context.Context, userID string) ([]*domain.FriendWithUser, error) {
	query := `
		SELECT f.id, f.requester_id, f.addressee_id, f.status, f.created_at, f.updated_at,
		       u.id, u.nickname, u.public_key, u.avatar_url, u.last_seen_at
		FROM friendships f
		JOIN users u ON f.requester_id = u.id
		WHERE f.addressee_id = $1 AND f.status = 'pending'
	`
	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*domain.FriendWithUser
	for rows.Next() {
		fw := &domain.FriendWithUser{Friend: &domain.User{}}
		err := rows.Scan(
			&fw.ID, &fw.RequesterID, &fw.AddresseeID, &fw.Status, &fw.CreatedAt, &fw.UpdatedAt,
			&fw.Friend.ID, &fw.Friend.Nickname, &fw.Friend.PublicKey, &fw.Friend.AvatarURL, &fw.Friend.LastSeenAt,
		)
		if err != nil {
			return nil, err
		}
		result = append(result, fw)
	}
	return result, rows.Err()
}

func (r *FriendshipRepository) UpdateStatus(ctx context.Context, id string, status domain.FriendshipStatus) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE friendships SET status = $2 WHERE id = $1`,
		id, status,
	)
	return err
}

func (r *FriendshipRepository) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM friendships WHERE id = $1`, id)
	return err
}

var _ domain.FriendshipRepository = (*FriendshipRepository)(nil)
