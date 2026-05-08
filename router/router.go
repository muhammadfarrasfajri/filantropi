package router

import (
	"github.com/gin-gonic/gin"
	"github.com/muhammadfarrasfajri/filantropi/controller"
	"github.com/muhammadfarrasfajri/filantropi/middleware"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func SetupRouter(r *gin.Engine, registerController *controller.RegisterController, loginController *controller.LoginController, refreshTokenController *controller.RefreshTokenController, userController *controller.UserController, jwtManager *middleware.JWTManager, campaignController *controller.CampaignController, donationController *controller.DonationController, adminController *controller.AdminController, webhookController *controller.WebhookController) {

	r.POST("/api/wallet/:wallet", userController.PostWallet)
	r.POST("/api/delete/:wallet", donationController.DeleteWallet)
	// r.GET("/api/wallet/:wallet", donationController.GetWalletStats)
	// r.GET("/api/donatur/:wallet", donationController.GetDonaturHistory)
	// r.GET("/api/campaign/:wallet/", donationController.GetTotalCollectedByCampaign)
	// routes.go
	r.POST("/api/webhooks/alchemy", webhookController.HandleAlchemyWebhook)
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
			privateCampaign.POST("/disbursements/:id", campaignController.RequestDisbursement)
			privateCampaign.POST("/report/:id", campaignController.SubmitReport)
			privateCampaign.GET("/milestone-status/:id", campaignController.GetMilestoneStatus)
			privateCampaign.GET("/stepper/:id", campaignController.GetCampaignStepper)

		}
	}

	donations := r.Group("/api/donations")
	{
		privateDonation := donations.Group("/")
		privateDonation.Use(jwtManager.AuthMiddleware())
		{
			privateDonation.GET("/history", donationController.MyHistory)
			// privateDonation.GET("/wallets/history/:wallet", donationController.GetWalletHistory)
			// privateDonation.GET("/wallet/balance/:wallet", donationController.GetCurrentBalance)
			privateDonation.GET("/amount/:wallet", donationController.GetTotalCollectedByCampaign)
			privateDonation.GET("/out/:wallet", donationController.GetDonaturHistory)
			privateDonation.GET("/in/:wallet", donationController.GetWalletStats)
		}
	}

	admins := r.Group("/api/admin/")
	{
		admins.POST("auth/login", loginController.LoginGoogleAdmin)
		privateAdmin := admins.Group("/")
		privateAdmin.Use(jwtManager.AuthMiddleware())
		privateAdmin.Use(middleware.AdminMiddleware())
		{
			privateAdmin.POST("auth/logout", refreshTokenController.LogoutAdmin)
			privateAdmin.GET("campaigns", adminController.GetAllCampaign)
			privateAdmin.GET("campaigns/:slug", campaignController.GetCampaignDetail)
			privateAdmin.PATCH("campaigns/:slug", adminController.ApproveCampaign)
			privateAdmin.GET("users", adminController.GetAllUsersForAdmin)
			privateAdmin.GET("users/:id", adminController.GetUserDetailForAdmin)
			privateAdmin.GET("dashboard", adminController.GetDashboardSummary)
			privateAdmin.PATCH("verified/:id", adminController.VerifyUser)
			privateAdmin.PATCH("campaigns/reject/:slug", adminController.RejectedCampaign)
			privateAdmin.PATCH("disbursements/approve/:id", adminController.ApproveDisbursement)
			privateAdmin.GET("disbursements/pending", campaignController.GetPendingDisbursements)
			privateAdmin.GET("reports/pending", campaignController.GetPendingReports)
			privateAdmin.PATCH("reports/approve/:id", campaignController.ApproveReport)
			privateAdmin.PATCH("reports/reject/:id", campaignController.RejectReport)
			privateAdmin.PATCH("update/wallet/:id", campaignController.UpdateWalletAddress)
		}
	}
}
