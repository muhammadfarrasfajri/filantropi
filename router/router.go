package router

import (
	"github.com/gin-gonic/gin"
	"github.com/muhammadfarrasfajri/filantropi/controller"
)

func SetupRouter(r *gin.Engine, userController *controller.RegisterController) {
	auth := r.Group("/api/auth")
	{
		auth.POST("/register", userController.RegisterGoogle)
	}
}
