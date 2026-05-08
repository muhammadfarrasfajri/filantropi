package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/muhammadfarrasfajri/filantropi/model"
)

func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		role := c.GetString("role")
		if role != "admin" {
			// Tolak akses jika bukan admin
			c.JSON(http.StatusForbidden, model.APIResponse{
				Error:   true,
				Message: "rejected access you are not admin",
				Type:    "Forbidden",
			})
			c.Abort() // SANGAT PENTING: Mencegah eksekusi handler/controller berikutnya
			return
		}

		// Lanjutkan ke handler/controller berikutnya jika lolos
		c.Next()
	}
}
