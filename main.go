package main

import (
	"github.com/gin-gonic/gin"
	"github.com/muhammadfarrasfajri/filantropi/bootstrap"
	_ "github.com/muhammadfarrasfajri/filantropi/docs" // Dokumentasi Swagger
	"github.com/muhammadfarrasfajri/filantropi/middleware"
	"github.com/muhammadfarrasfajri/filantropi/router"
)

// @title           Filantropi API
// @version         1.0
// @description     Backend API untuk platform filantropi.
// @termsOfService  http://swagger.io/terms/

// @contact.name    Muhammad Farras Fajri
// @contact.url     https://github.com/muhammadfarrasfajri

// @license.name    Apache 2.0
// @license.url     http://www.apache.org/licenses/LICENSE-2.0.html

// @host            localhost:8080
// @BasePath        /

// @securityDefinitions.apikey BearerAuth
// @in                         header
// @name                       Authorization
// @description                Masukkan token dengan format: "Bearer <your_token>"
func main() {
	// 1. Inisialisasi Database
	bootstrap.InitDatabase()

	// 2. Inisialisasi Firebase (untuk Google Auth)
	firebaseApp := bootstrap.InitFirebase()

	// 3. Inisialisasi Dependency Injection (Container)
	container := bootstrap.InitContainer(firebaseApp)

	// 4. Inisialisasi Gin Engine
	r := gin.Default()

	// 6. Global Middleware
	middleware.AttachCORS(r)

	// 5. Konfigurasi Static Files & Upload Memory
	// Mengizinkan akses file di folder public (seperti foto profil)
	r.Static("/public", "./public")
	// Limit memory untuk upload file (50 MiB)
	r.MaxMultipartMemory = 50 << 20

	// 7. Setup Routes
	// Melewatkan semua controller dari container ke router
	router.SetupRouter(
		r,
		container.RegisterController,
		container.LoginController,
		container.RefreshController,
		container.UserController,
		container.JWTManager,
		container.CampaignController,
		container.DonationController,
	)

	// 8. Jalankan Server
	r.Run(":8080")
}
