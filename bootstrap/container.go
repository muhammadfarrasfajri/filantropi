package bootstrap

import (
	"firebase.google.com/go/auth"
	"github.com/muhammadfarrasfajri/filantropi/controller"
	"github.com/muhammadfarrasfajri/filantropi/database"
	"github.com/muhammadfarrasfajri/filantropi/middleware"
	"github.com/muhammadfarrasfajri/filantropi/repository"
	"github.com/muhammadfarrasfajri/filantropi/service"
)

type Container struct {
	RegisterController *controller.RegisterController
	LoginController    *controller.LoginController
	RefreshController  *controller.RefreshTokenController
	UserController     *controller.UserController
	JWTManager         *middleware.JWTManager
	CampaignController *controller.CampaignController
	DonationController *controller.DonationController
}

func InitContainer(userAuth *auth.Client) *Container {
	// Initialize repositories
	registerRepo := repository.NewRegisterRepository(database.DB)
	loginRepo := repository.NewLoginRepository(database.DB)
	refreshRepo := repository.NewRefreshTokenRepository(database.DB)
	userRepo := repository.NewUserRepository(database.DB)
	campaignRepo := repository.NewCampaignRepository(database.DB)
	donationRepo := repository.NewDonationRepository(database.DB)

	// Initialize middleware
	jwtManager := middleware.NewJWTManager("79329e633bbbd5652893feea5c27f60faa0ee69688e65e29bf03419889be965adcab16420e07fa88c62ab8d1f7c82804aee66e30d237f6381b002e1ae1109187", "aaa46ab1983939bfaa571d4b6581e2012d0cec9a67ee1cec975f64af716f6850f080d389b12b70b6918e94bf417c7448133cbd129c8c4bf567bdb7f82bbfa3a1")

	// Initialize services
	registerService := service.NewRegistrasiService(registerRepo, refreshRepo, jwtManager, userAuth)
	loginService := service.NewLoginService(loginRepo, refreshRepo, jwtManager, userAuth)
	refreshService := service.NewRefreshTokenService(refreshRepo, userRepo, jwtManager)
	userService := service.NewUserService(userRepo, registerRepo)
	campaignService := service.NewCampaignService(campaignRepo, userRepo)
	donationService := service.NewDonationService(donationRepo, campaignRepo, userRepo)

	// Initialize controllers
	registerController := controller.NewRegisterController(registerService)
	loginController := controller.NewLoginController(loginService)
	refreshController := controller.NewRefreshTokenController(refreshService)
	userController := controller.NewUserController(userService)
	campaignController := controller.NewCampaignController(campaignService)
	donationController := controller.NewDonationController(donationService)

	return &Container{
		RegisterController: registerController,
		LoginController:    loginController,
		RefreshController:  refreshController,
		UserController:     userController,
		JWTManager:         jwtManager,
		CampaignController: campaignController,
		DonationController: donationController,
	}
}
