package handler

import (
	"link/internal/domain"
	"link/internal/service"

	"github.com/gofiber/fiber/v2"
)

type FriendHandler struct {
	friendSvc *service.FriendshipService
}

func NewFriendHandler(friendSvc *service.FriendshipService) *FriendHandler {
	return &FriendHandler{friendSvc: friendSvc}
}

func (h *FriendHandler) List(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	friends, err := h.friendSvc.GetFriends(c.Context(), userID)
	if err != nil {
		return Error(c, err)
	}
	return OK(c, friends)
}

func (h *FriendHandler) Requests(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	requests, err := h.friendSvc.GetPendingRequests(c.Context(), userID)
	if err != nil {
		return Error(c, err)
	}
	return OK(c, requests)
}

func (h *FriendHandler) SendRequest(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	var req struct {
		AddresseeID string `json:"addressee_id"`
	}
	if err := c.BodyParser(&req); err != nil {
		return Error(c, domain.ErrValidation("invalid request"))
	}

	friendship, err := h.friendSvc.SendRequest(c.Context(), userID, req.AddresseeID)
	if err != nil {
		return Error(c, err)
	}
	return OK(c, friendship)
}

func (h *FriendHandler) Accept(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	friendshipID := c.Params("id")

	if err := h.friendSvc.Accept(c.Context(), userID, friendshipID); err != nil {
		return Error(c, err)
	}
	return OK(c, fiber.Map{"message": "已接受好友請求"})
}

func (h *FriendHandler) Reject(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	friendshipID := c.Params("id")

	if err := h.friendSvc.Reject(c.Context(), userID, friendshipID); err != nil {
		return Error(c, err)
	}
	return OK(c, fiber.Map{"message": "已拒絕好友請求"})
}

func (h *FriendHandler) Remove(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	friendshipID := c.Params("id")

	if err := h.friendSvc.Remove(c.Context(), userID, friendshipID); err != nil {
		return Error(c, err)
	}
	return OK(c, fiber.Map{"message": "已移除好友"})
}
