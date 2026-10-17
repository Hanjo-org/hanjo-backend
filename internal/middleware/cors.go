package middleware

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
)

// CORS configures cross-origin resource sharing.
func CORS(allowedOrigins []string) fiber.Handler {
	if len(allowedOrigins) == 0 {
		return func(c fiber.Ctx) error {
			return c.Next()
		}
	}

	return cors.New(cors.Config{
		AllowOrigins:     allowedOrigins,
		AllowCredentials: true,
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With"},
		AllowMethods:     []string{"GET", "POST", "HEAD", "PUT", "DELETE", "PATCH", "OPTIONS"},
	})
}
