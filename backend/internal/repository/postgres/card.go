package postgres

import (
	"context"

	"link/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CardRepository struct {
	pool *pgxpool.Pool
}

func NewCardRepository(pool *pgxpool.Pool) *CardRepository {
	return &CardRepository{pool: pool}
}

func (r *CardRepository) FindByToken(ctx context.Context, token string) (*domain.Card, error) {
	query := `
		SELECT id, user_id, card_token, card_type, status, created_at, activated_at, revoked_at
		FROM cards WHERE card_token = $1
	`
	card := &domain.Card{}
	err := r.pool.QueryRow(ctx, query, token).Scan(
		&card.ID, &card.UserID, &card.CardToken, &card.CardType,
		&card.Status, &card.CreatedAt, &card.ActivatedAt, &card.RevokedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return card, nil
}

func (r *CardRepository) FindByUserID(ctx context.Context, userID string) ([]*domain.Card, error) {
	query := `
		SELECT id, user_id, card_token, card_type, status, created_at, activated_at, revoked_at
		FROM cards WHERE user_id = $1
	`
	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cards []*domain.Card
	for rows.Next() {
		c := &domain.Card{}
		if err := rows.Scan(&c.ID, &c.UserID, &c.CardToken, &c.CardType,
			&c.Status, &c.CreatedAt, &c.ActivatedAt, &c.RevokedAt); err != nil {
			return nil, err
		}
		cards = append(cards, c)
	}
	return cards, rows.Err()
}

func (r *CardRepository) FindActiveByUserAndType(ctx context.Context, userID string, cardType domain.CardType) (*domain.Card, error) {
	query := `
		SELECT id, user_id, card_token, card_type, status, created_at, activated_at, revoked_at
		FROM cards WHERE user_id = $1 AND card_type = $2 AND status = 'active'
	`
	card := &domain.Card{}
	err := r.pool.QueryRow(ctx, query, userID, cardType).Scan(
		&card.ID, &card.UserID, &card.CardToken, &card.CardType,
		&card.Status, &card.CreatedAt, &card.ActivatedAt, &card.RevokedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return card, err
}

func (r *CardRepository) Create(ctx context.Context, card *domain.Card) error {
	query := `
		INSERT INTO cards (user_id, card_token, card_type, status)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at
	`
	return r.pool.QueryRow(ctx, query,
		card.UserID, card.CardToken, card.CardType, card.Status,
	).Scan(&card.ID, &card.CreatedAt)
}

func (r *CardRepository) Revoke(ctx context.Context, cardID string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE cards SET status = 'revoked', revoked_at = NOW() WHERE id = $1`,
		cardID,
	)
	return err
}

func (r *CardRepository) PromoteBackupToPrimary(ctx context.Context, cardID string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE cards SET card_type = 'primary', activated_at = NOW() WHERE id = $1`,
		cardID,
	)
	return err
}

func (r *CardRepository) CreatePair(ctx context.Context, primaryToken string) (*domain.CardPair, error) {
	pair := &domain.CardPair{}
	query := `
		INSERT INTO card_pairs (primary_token)
		VALUES ($1)
		RETURNING id, primary_token, backup_token, created_at, expires_at
	`
	err := r.pool.QueryRow(ctx, query, primaryToken).Scan(
		&pair.ID, &pair.PrimaryToken, &pair.BackupToken, &pair.CreatedAt, &pair.ExpiresAt,
	)
	return pair, err
}

func (r *CardRepository) FindPairByPrimaryToken(ctx context.Context, token string) (*domain.CardPair, error) {
	query := `
		SELECT id, primary_token, backup_token, created_at, expires_at
		FROM card_pairs WHERE primary_token = $1 AND expires_at > NOW()
	`
	pair := &domain.CardPair{}
	err := r.pool.QueryRow(ctx, query, token).Scan(
		&pair.ID, &pair.PrimaryToken, &pair.BackupToken, &pair.CreatedAt, &pair.ExpiresAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return pair, err
}

func (r *CardRepository) FindPairByBackupToken(ctx context.Context, token string) (*domain.CardPair, error) {
	query := `
		SELECT id, primary_token, backup_token, created_at, expires_at
		FROM card_pairs WHERE backup_token = $1 AND expires_at > NOW()
	`
	pair := &domain.CardPair{}
	err := r.pool.QueryRow(ctx, query, token).Scan(
		&pair.ID, &pair.PrimaryToken, &pair.BackupToken, &pair.CreatedAt, &pair.ExpiresAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return pair, err
}

func (r *CardRepository) UpdatePairBackupToken(ctx context.Context, pairID, backupToken string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE card_pairs SET backup_token = $2 WHERE id = $1`,
		pairID, backupToken,
	)
	return err
}

func (r *CardRepository) DeletePair(ctx context.Context, pairID string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM card_pairs WHERE id = $1`, pairID)
	return err
}

func (r *CardRepository) CleanupExpiredPairs(ctx context.Context) error {
	_, err := r.pool.Exec(ctx, `SELECT cleanup_expired_pairs()`)
	return err
}

var _ domain.CardRepository = (*CardRepository)(nil)
