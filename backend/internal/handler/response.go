package handler

import (
	"log/slog"

	"link/internal/domain"

	"github.com/gofiber/fiber/v2"
)

func OK(c *fiber.Ctx, data interface{}) error {
	return c.JSON(fiber.Map{"data": data})
}

func Error(c *fiber.Ctx, err error) error {
	if appErr, ok := domain.IsAppError(err); ok {
		return c.Status(appErr.Status).JSON(fiber.Map{
			"error": fiber.Map{"code": appErr.Code, "message": appErr.Message},
		})
	}
	slog.Error("unhandled error", "err", err)
	return c.Status(500).JSON(fiber.Map{
		"error": fiber.Map{"code": domain.ErrCodeInternal, "message": "系統錯誤"},
	})
}
