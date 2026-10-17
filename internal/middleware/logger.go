package middleware

import (
	"time"

	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
)

// ZapLogger creates a structured HTTP request logging middleware using Zap.
func ZapLogger(log *zap.Logger) fiber.Handler {
	return func(c fiber.Ctx) error {
		start := time.Now()
		err := c.Next()
		latency := time.Since(start)

		status := c.Response().StatusCode()
		fields := []zap.Field{
			zap.Int("status", status),
			zap.String("method", c.Method()),
			zap.String("path", c.Path()),
			zap.String("ip", c.IP()),
			zap.Duration("latency", latency),
		}

		if err != nil {
			fields = append(fields, zap.Error(err))
		}

		if status >= 500 {
			log.Error("http request error", fields...)
		} else if status >= 400 {
			log.Warn("http request warning", fields...)
		} else {
			log.Info("http request", fields...)
		}

		return err
	}
}
