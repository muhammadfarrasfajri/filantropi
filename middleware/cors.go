package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func AttachCORS(r *gin.Engine) {
	// List domain yang boleh akses API kamu
	allowedOrigins := map[string]bool{
		"http://localhost:3000":      true,
		"http://localhost:3001":      true,
		"https://wanama.id":          true, // Ganti dengan domain production nanti
		"http://192.168.52.211:3000": true,
		"http://192.168.52.211:3001": true,
	}

	r.Use(func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// Cek apakah origin ada di whitelist
		if allowedOrigins[origin] {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		} else if origin == "" {
			// Fallback untuk request non-browser (misal Postman/Curl)
			c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		}

		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, PATCH")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")

		// Beri tahu browser untuk tidak selalu menanyakan izin CORS (Caching)
		c.Writer.Header().Set("Access-Control-Max-Age", "86400")

		// Beri tahu browser bahwa response bisa berubah tergantung Origin
		c.Writer.Header().Set("Vary", "Origin")

		// Handle Preflight Request
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	})
}
