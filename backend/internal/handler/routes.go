package handler

import (
	"time"

	"link/internal/middleware"

	"github.com/gofiber/fiber/v2"
)

type Handlers struct {
	Auth   *AuthHandler
	User   *UserHandler
	Friend *FriendHandler
	Conv   *ConversationHandler
	Admin  *AdminHandler
}

func Setup(app *fiber.App, h *Handlers, authMw fiber.Handler) {
	app.Get("/health", func(c *fiber.Ctx) error { return c.SendString("OK") })

	api := app.Group("/api/v1")

	loginLimiter := middleware.NewRateLimiter(10, time.Minute)
	registerLimiter := middleware.NewRateLimiter(5, time.Hour)

	api.Get("/auth/check-card/:token", h.Auth.CheckCard)
	api.Post("/auth/register", registerLimiter.Middleware(), h.Auth.Register)
	api.Post("/auth/login", loginLimiter.Middleware(), h.Auth.Login)
	api.Post("/auth/login/backup", loginLimiter.Middleware(), h.Auth.LoginWithBackup)
	app.Get("/w/:token", h.Auth.CardEntry)

	// Admin routes (before auth middleware group)
	admin := api.Group("/admin", h.Admin.AuthMiddleware())
	admin.Post("/cards/generate", h.Admin.GenerateCardPair)
	admin.Get("/cards", h.Admin.ListCardPairs)
	admin.Delete("/cards/:id", h.Admin.DeleteCardPair)

	auth := api.Group("", authMw)
	auth.Get("/users/me", h.User.GetMe)
	auth.Get("/users/me/cards", h.User.GetMyCards)
	auth.Patch("/users/me", h.User.UpdateMe)
	auth.Get("/users/search", h.User.Search)
	auth.Get("/users/:id/public-key", h.User.GetPublicKey)

	auth.Get("/friends", h.Friend.List)
	auth.Get("/friends/requests", h.Friend.Requests)
	auth.Post("/friends/request", h.Friend.SendRequest)
	auth.Post("/friends/:id/accept", h.Friend.Accept)
	auth.Post("/friends/:id/reject", h.Friend.Reject)
	auth.Delete("/friends/:id", h.Friend.Remove)

	auth.Get("/conversations", h.Conv.List)
	auth.Get("/conversations/:id/messages", h.Conv.Messages)
	auth.Delete("/messages/:messageId", h.Conv.DeleteMessage)

	auth.Post("/auth/logout", h.Auth.Logout)
}
