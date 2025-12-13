package handler

import (
	"link/internal/domain"
	"link/internal/service"

	"github.com/gofiber/fiber/v2"
)

type UserHandler struct {
	userSvc *service.UserService
	cardSvc *service.CardService
}

func NewUserHandler(userSvc *service.UserService, cardSvc *service.CardService) *UserHandler {
	return &UserHandler{userSvc: userSvc, cardSvc: cardSvc}
}

func (h *UserHandler) GetMe(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	user, err := h.userSvc.GetByID(c.Context(), userID)
	if err != nil {
		return Error(c, err)
	}
	return OK(c, user)
}

func (h *UserHandler) GetMyCards(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	cards, err := h.cardSvc.GetUserCards(c.Context(), userID)
	if err != nil {
		return Error(c, err)
	}
	return OK(c, cards)
}

func (h *UserHandler) UpdateMe(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	var req struct {
		Nickname  *string `json:"nickname"`
		AvatarURL *string `json:"avatar_url"`
		PublicKey *string `json:"public_key"`
	}
	if err := c.BodyParser(&req); err != nil {
		return Error(c, domain.ErrValidation("invalid request"))
	}

	user, err := h.userSvc.GetByID(c.Context(), userID)
	if err != nil {
		return Error(c, err)
	}

	if req.Nickname != nil {
		user.Nickname = *req.Nickname
	}
	if req.AvatarURL != nil {
		user.AvatarURL = req.AvatarURL
	}
	if req.PublicKey != nil {
		user.PublicKey = *req.PublicKey
	}

	if err := h.userSvc.Update(c.Context(), user); err != nil {
		return Error(c, err)
	}
	return OK(c, user)
}

func (h *UserHandler) Search(c *fiber.Ctx) error {
	query := c.Query("q")
	if query == "" {
		return Error(c, domain.ErrValidation("query required"))
	}
	users, err := h.userSvc.Search(c.Context(), query, 20)
	if err != nil {
		return Error(c, err)
	}
	return OK(c, users)
}

func (h *UserHandler) GetPublicKey(c *fiber.Ctx) error {
	id := c.Params("id")
	pk, err := h.userSvc.GetPublicKey(c.Context(), id)
	if err != nil {
		return Error(c, err)
	}
	return OK(c, fiber.Map{"public_key": pk})
}
