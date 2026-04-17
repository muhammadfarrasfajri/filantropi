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
	UserController *controller.RegisterController
}

func InitContainer(userAuth *auth.Client) *Container {
	// Initialize repositories
	registerRepo := repository.NewRegisterRepository(database.DB)
	refreshRepo := repository.NewRefreshTokenRepository(database.DB)

	// Initialize middleware
	jwtManager := middleware.NewJWTManager("79329e633bbbd5652893feea5c27f60faa0ee69688e65e29bf03419889be965adcab16420e07fa88c62ab8d1f7c82804aee66e30d237f6381b002e1ae1109187", "aaa46ab1983939bfaa571d4b6581e2012d0cec9a67ee1cec975f64af716f6850f080d389b12b70b6918e94bf417c7448133cbd129c8c4bf567bdb7f82bbfa3a1")

	// Initialize services
	registerService := service.NewRegistrasiService(registerRepo, refreshRepo, jwtManager, userAuth)

	// Initialize controllers
	userController := controller.NewRegisterController(registerService)

	return &Container{
		UserController: userController,
	}
}
