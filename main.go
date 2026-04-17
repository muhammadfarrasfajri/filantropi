package main

import (
	"github.com/gin-gonic/gin"
	"github.com/muhammadfarrasfajri/filantropi/bootstrap"
	"github.com/muhammadfarrasfajri/filantropi/middleware"
	"github.com/muhammadfarrasfajri/filantropi/router"
)

func main() {

	// Database initialization
	bootstrap.InitDatabase()

	// Firebase initialization
	firebase := bootstrap.InitFirebase()

	// Container initialization
	container := bootstrap.InitContainer(firebase)

	// Gin
	r := gin.Default()

	r.Static("/public", "./public")

	r.MaxMultipartMemory = 50 << 20

	// CORS Middleware
	middleware.AttachCORS(r)

	// ROUTES
	router.SetupRouter(r, container.UserController)

	r.Run(":8080")
}
