package middleware

import (
	"sync"
	"time"

	"link/internal/domain"

	"github.com/gofiber/fiber/v2"
)

type RateLimiter struct {
	visitors map[string]*visitor
	mu       sync.RWMutex
	rate     int
	window   time.Duration
}

type visitor struct {
	count    int
	lastSeen time.Time
}

func NewRateLimiter(rate int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*visitor),
		rate:     rate,
		window:   window,
	}
	go rl.cleanup()
	return rl
}

func (rl *RateLimiter) cleanup() {
	for {
		time.Sleep(rl.window)
		rl.mu.Lock()
		for ip, v := range rl.visitors {
			if time.Since(v.lastSeen) > rl.window {
				delete(rl.visitors, ip)
			}
		}
		rl.mu.Unlock()
	}
}

func (rl *RateLimiter) Middleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		ip := c.IP()

		rl.mu.Lock()
		v, exists := rl.visitors[ip]
		if !exists || time.Since(v.lastSeen) > rl.window {
			rl.visitors[ip] = &visitor{count: 1, lastSeen: time.Now()}
			rl.mu.Unlock()
			return c.Next()
		}

		v.count++
		v.lastSeen = time.Now()

		if v.count > rl.rate {
			rl.mu.Unlock()
			appErr := domain.ErrRateLimited()
			return c.Status(appErr.Status).JSON(fiber.Map{
				"error": fiber.Map{"code": appErr.Code, "message": appErr.Message},
			})
		}
		rl.mu.Unlock()

		return c.Next()
	}
}
