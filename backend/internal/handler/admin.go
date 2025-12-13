package handler

import (
	"sync"

	"link/internal/pkg/cardtoken"

	"github.com/gofiber/fiber/v2"
)

type AdminHandler struct {
	tokenGen   *cardtoken.Generator
	password   string
	baseURL    string
	cardPairs  []CardPairInfo
	pairsMutex sync.RWMutex
	nextID     int
}

type CardPairInfo struct {
	ID         int    `json:"id"`
	FirstToken string `json:"first_token"`
	SecondToken string `json:"second_token"`
	FirstURL   string `json:"first_url"`
	SecondURL  string `json:"second_url"`
}

func NewAdminHandler(tokenGen *cardtoken.Generator, password, baseURL string) *AdminHandler {
	return &AdminHandler{
		tokenGen:  tokenGen,
		password:  password,
		baseURL:   baseURL,
		cardPairs: make([]CardPairInfo, 0),
		nextID:    1,
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

	pair := CardPairInfo{
		ID:          h.nextID,
		FirstToken:  first,
		SecondToken: second,
		FirstURL:    "https://192.168.1.99:9443/w/" + first,
		SecondURL:   "https://192.168.1.99:9443/w/" + second,
	}

	h.pairsMutex.Lock()
	h.cardPairs = append(h.cardPairs, pair)
	h.nextID++
	h.pairsMutex.Unlock()

	return c.JSON(fiber.Map{"data": pair})
}

func (h *AdminHandler) ListCardPairs(c *fiber.Ctx) error {
	h.pairsMutex.RLock()
	defer h.pairsMutex.RUnlock()

	return c.JSON(fiber.Map{"data": h.cardPairs})
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
