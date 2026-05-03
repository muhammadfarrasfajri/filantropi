package router

import (
	"github.com/gin-gonic/gin"
	"github.com/muhammadfarrasfajri/filantropi/controller"
	"github.com/muhammadfarrasfajri/filantropi/middleware"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func SetupRouter(r *gin.Engine, registerController *controller.RegisterController, loginController *controller.LoginController, refreshTokenController *controller.RefreshTokenController, userController *controller.UserController, jwtManager *middleware.JWTManager, campaignController *controller.CampaignController, donationController *controller.DonationController) {

	// Swagger documentation
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Group API Auth
	auth := r.Group("/api/auth")
	{
		auth.POST("/register/donor", registerController.RegisterGoogleUser)
		auth.POST("/register/beneficiary", registerController.RegisterGoogleBeneficiaries)
		auth.POST("/login", loginController.LoginGoogle)
		auth.POST("/refresh-token", refreshTokenController.RefreshToken)

		// Hanya logout yang butuh middleware di group ini
		auth.POST("/logout", jwtManager.AuthMiddleware(), refreshTokenController.Logout)
	}

	// Group API User - Semua rute di sini butuh Login
	user := r.Group("/api/user")
	user.Use(jwtManager.AuthMiddleware())
	{
		user.GET("/profile/donors", userController.FindById)
		user.GET("/profile/beneficiaries", userController.FindBeneficiaryById)
		user.PUT("/profile/update-donors", userController.UpdateDonors)
		user.PUT("/profile/update-beneficiaries", userController.UpdateProfileBeneficiaries)
	}

	// Group API Campaigns
	campaigns := r.Group("/api/campaigns")
	{
		// Rute Publik (Bisa diakses tanpa login)
		campaigns.GET("/", campaignController.GetActiveCampaigns)
		campaigns.GET("/:slug", campaignController.GetCampaignDetail)

		// Rute Privat (Wajib Login)
		privateCampaign := campaigns.Group("/")
		privateCampaign.Use(jwtManager.AuthMiddleware())
		{
			privateCampaign.POST("", campaignController.CreateCampaign)
			privateCampaign.GET("/me", campaignController.GetMyCampaigns)
			privateCampaign.PUT("/:slug", campaignController.UpdateCampaign)
			privateCampaign.PATCH("/admin/approve/:id", campaignController.ApproveCampaign)
		}
	}

	donations := r.Group("/api/donations")
	{
		privateDonation := donations.Group("/")
		privateDonation.Use(jwtManager.AuthMiddleware())
		{
			privateDonation.POST("", donationController.CreateDonation)
			privateDonation.GET("/history", donationController.MyHistory)
			privateDonation.GET("/wallets/history/:wallet", donationController.GetWalletHistory)
		}
	}
}
