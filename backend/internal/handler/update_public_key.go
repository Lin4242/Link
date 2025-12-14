package handler

import (
	"link/internal/domain"
	"github.com/gofiber/fiber/v2"
)

func (h *UserHandler) UpdatePublicKey(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	
	var req struct {
		PublicKey string `json:"public_key"`
	}
	
	if err := c.BodyParser(&req); err != nil {
		return Error(c, domain.ErrValidation("invalid request"))
	}
	
	if req.PublicKey == "" {
		return Error(c, domain.ErrValidation("public key is required"))
	}
	
	user, err := h.userSvc.GetByID(c.Context(), userID)
	if err != nil {
		return Error(c, err)
	}
	
	user.PublicKey = req.PublicKey
	
	if err := h.userSvc.Update(c.Context(), user); err != nil {
		return Error(c, err)
	}
	
	return OK(c, fiber.Map{
		"message": "Public key updated successfully",
		"public_key": req.PublicKey,
	})
}