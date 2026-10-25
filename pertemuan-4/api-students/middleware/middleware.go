package middleware

import (
	"log/slog"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"api-students/helper"
)

// Register memasang seluruh middleware global yang berlaku untuk semua route.
// URUTAN PENTING: middleware dieksekusi secara berantai sesuai urutan pemasangan.
func Register(app *fiber.App, logger *slog.Logger) {
	app.Use(requestid.New())       // 1. Beri setiap request satu ID unik (X-Request-Id)
	app.Use(recover.New())         // 2. Tangkap panic agar server tidak crash
	app.Use(helmet.New())          // 3. Pasang header proteksi keamanan dasar (Security Headers)
	app.Use(cors.New())            // 4. Konfigurasi Cross-Origin Resource Sharing
	app.Use(RequestLogger(logger)) // 5. Catat setiap HTTP request ke log terstruktur
}

// RequestLogger mencatat setiap HTTP request ke sistem log terstruktur JSON.
func RequestLogger(logger *slog.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next()

		requestID, _ := c.Locals("requestid").(string)
		logger.Info("http_request",
			slog.String("request_id", requestID),
			slog.String("method", c.Method()),
			slog.String("path", c.Path()),
			slog.Int("status", c.Response().StatusCode()),
			slog.Duration("duration", time.Since(start)),
			slog.String("ip", c.IP()),
		)
		return err
	}
}

var methodsWithBody = map[string]bool{
	fiber.MethodPost:  true,
	fiber.MethodPut:   true,
	fiber.MethodPatch: true,
}

// RequireJSON menolak permintaan yang membawa body dengan Content-Type selain application/json.
// Mengembalikan status 415 Unsupported Media Type.
func RequireJSON(c *fiber.Ctx) error {
	if methodsWithBody[c.Method()] {
		ct := c.Get("Content-Type")
		if !strings.HasPrefix(ct, fiber.MIMEApplicationJSON) {
			return helper.Fail(c, fiber.StatusUnsupportedMediaType, "Content-Type harus application/json")
		}
	}
	return c.Next()
}
