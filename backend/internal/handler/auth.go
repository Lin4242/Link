package handler

import (
	"link/internal/domain"
	"link/internal/service"

	"github.com/gofiber/fiber/v2"
)

type AuthHandler struct {
	authSvc *service.AuthService
	cardSvc *service.CardService
	baseURL string
}

func NewAuthHandler(authSvc *service.AuthService, cardSvc *service.CardService, baseURL string) *AuthHandler {
	return &AuthHandler{authSvc: authSvc, cardSvc: cardSvc, baseURL: baseURL}
}

func (h *AuthHandler) CheckCard(c *fiber.Ctx) error {
	token := c.Params("token")
	result, err := h.cardSvc.CheckCard(c.Context(), token)
	if err != nil {
		return Error(c, err)
	}
	return OK(c, result)
}

func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req struct {
		PrimaryToken string `json:"primary_token"`
		BackupToken  string `json:"backup_token"`
		Password     string `json:"password"`
		Nickname     string `json:"nickname"`
		PublicKey    string `json:"public_key"`
	}
	if err := c.BodyParser(&req); err != nil {
		return Error(c, domain.ErrValidation("invalid request"))
	}

	res, err := h.authSvc.Register(c.Context(), service.RegisterInput{
		PrimaryToken: req.PrimaryToken,
		BackupToken:  req.BackupToken,
		Password:     req.Password,
		Nickname:     req.Nickname,
		PublicKey:    req.PublicKey,
	})
	if err != nil {
		return Error(c, err)
	}
	return OK(c, res)
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req struct {
		CardToken string `json:"card_token"`
		Password  string `json:"password"`
	}
	if err := c.BodyParser(&req); err != nil {
		return Error(c, domain.ErrValidation("invalid request"))
	}

	res, err := h.authSvc.Login(c.Context(), req.CardToken, req.Password)
	if err != nil {
		return Error(c, err)
	}
	return OK(c, res)
}

func (h *AuthHandler) LoginWithBackup(c *fiber.Ctx) error {
	var req struct {
		CardToken string `json:"card_token"`
		Password  string `json:"password"`
		Confirm   bool   `json:"confirm"`
	}
	if err := c.BodyParser(&req); err != nil {
		return Error(c, domain.ErrValidation("invalid request"))
	}

	if !req.Confirm {
		return Error(c, domain.ErrValidation("必須確認撤銷主卡"))
	}

	res, err := h.authSvc.LoginWithBackupCard(c.Context(), req.CardToken, req.Password)
	if err != nil {
		return Error(c, err)
	}
	return OK(c, res)
}

func (h *AuthHandler) CardEntry(c *fiber.Ctx) error {
	token := c.Params("token")
	result, _ := h.cardSvc.CheckCard(c.Context(), token)

	switch result.Status {
	case "can_register":
		return c.Redirect(h.baseURL + "/register?token=" + token)
	case "primary":
		return c.Redirect(h.baseURL + "/login?token=" + token)
	case "backup":
		return c.Redirect(h.baseURL + "/login/backup?token=" + token)
	case "revoked":
		return c.Redirect(h.baseURL + "/error?reason=card_revoked")
	case "invalid_token":
		return c.Redirect(h.baseURL + "/error?reason=invalid_token")
	case "pair_already_registered":
		return c.Redirect(h.baseURL + "/error?reason=pair_registered")
	default:
		return c.Redirect(h.baseURL + "/error")
	}
}

func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	return OK(c, fiber.Map{"message": "已登出"})
}
