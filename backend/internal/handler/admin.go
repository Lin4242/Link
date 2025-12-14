package handler

import (
	"context"
	"sync"
	"time"

	"link/internal/pkg/cardtoken"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AdminHandler struct {
	tokenGen   *cardtoken.Generator
	password   string
	baseURL    string
	cardPairs  []CardPairInfo
	pairsMutex sync.RWMutex
	nextID     int
	db         *pgxpool.Pool
}

type CardPairInfo struct {
	ID         int    `json:"id"`
	FirstToken string `json:"first_token"`
	SecondToken string `json:"second_token"`
	FirstURL   string `json:"first_url"`
	SecondURL  string `json:"second_url"`
	IsActivated bool  `json:"is_activated"`
}

func NewAdminHandler(tokenGen *cardtoken.Generator, password, baseURL string, db *pgxpool.Pool) *AdminHandler {
	return &AdminHandler{
		tokenGen:  tokenGen,
		password:  password,
		baseURL:   baseURL,
		cardPairs: make([]CardPairInfo, 0),
		nextID:    1,
		db:        db,
	}
}

func (h *AdminHandler) AuthMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		password := c.Get("X-Admin-Password")
		if password != h.password {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid password",
			})
		}
		return c.Next()
	}
}

func (h *AdminHandler) GenerateCardPair(c *fiber.Ctx) error {
	first, second, err := h.tokenGen.GeneratePair()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to generate card pair",
		})
	}

	// Save to database with 30 days expiry
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = h.db.Exec(ctx,
		"INSERT INTO card_pairs (primary_token, backup_token, expires_at) VALUES ($1, $2, NOW() + INTERVAL '30 days')",
		first, second)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to save card pair",
		})
	}

	pair := CardPairInfo{
		ID:          h.nextID,
		FirstToken:  first,
		SecondToken: second,
		FirstURL:    h.baseURL + "/w/" + first,
		SecondURL:   h.baseURL + "/w/" + second,
	}

	h.pairsMutex.Lock()
	h.cardPairs = append(h.cardPairs, pair)
	h.nextID++
	h.pairsMutex.Unlock()

	return c.JSON(fiber.Map{"data": pair})
}

func (h *AdminHandler) ListCardPairs(c *fiber.Ctx) error {
	// Get from database
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := h.db.Query(ctx, `
		SELECT 
			cp.primary_token, 
			cp.backup_token, 
			cp.created_at,
			CASE WHEN c.id IS NOT NULL THEN true ELSE false END as is_activated
		FROM card_pairs cp
		LEFT JOIN cards c ON c.card_token = cp.primary_token AND c.card_type = 'primary' AND c.status = 'active'
		WHERE cp.expires_at > NOW() 
		ORDER BY cp.created_at ASC
	`)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to fetch card pairs",
		})
	}
	defer rows.Close()

	var pairs []CardPairInfo
	id := 1
	for rows.Next() {
		var primaryToken, backupToken string
		var createdAt time.Time
		var isActivated bool
		if err := rows.Scan(&primaryToken, &backupToken, &createdAt, &isActivated); err != nil {
			continue
		}

		pairs = append(pairs, CardPairInfo{
			ID:          id,
			FirstToken:  primaryToken,
			SecondToken: backupToken,
			FirstURL:    h.baseURL + "/w/" + primaryToken,
			SecondURL:   h.baseURL + "/w/" + backupToken,
			IsActivated: isActivated,
		})
		id++
	}

	return c.JSON(fiber.Map{"data": pairs})
}

func (h *AdminHandler) DeleteCardPair(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid id",
		})
	}

	h.pairsMutex.Lock()
	defer h.pairsMutex.Unlock()

	for i, pair := range h.cardPairs {
		if pair.ID == id {
			h.cardPairs = append(h.cardPairs[:i], h.cardPairs[i+1:]...)
			return c.JSON(fiber.Map{"message": "deleted"})
		}
	}

	return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
		"error": "not found",
	})
}
