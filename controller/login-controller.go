package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/muhammadfarrasfajri/filantropi/model"
	"github.com/muhammadfarrasfajri/filantropi/service"
)

type LoginController struct {
	LoginService *service.LoginService
}

func NewLoginController(loginService *service.LoginService) *LoginController {
	return &LoginController{
		LoginService: loginService,
	}
}

// @Summary      Login with Google
// @Description  Autentikasi user menggunakan IdToken dari Google. Jika berhasil, akan mengembalikan Access Token dan Refresh Token.
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Param        request  body      model.User  true  "Google ID Token"
// @Success      200      {object}  model.APIResponse{data=object} "Login Berhasil"
// @Failure      400      {object}  model.APIResponse "Format data salah"
// @Failure      401      {object}  model.APIResponse "Token Google tidak valid"
// @Router       /api/auth/login [post]
func (c LoginController) LoginGoogle(ctx *gin.Context) {
	var req model.User

	err := ctx.ShouldBind(&req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, model.APIResponse{
			Error:   true,
			Message: "Invalid data format",
			Type:    "ValidationError",
		})
		return
	}

	if req.IdToken == "" {
		ctx.JSON(http.StatusBadRequest, model.APIResponse{
			Error:   true,
			Message: "Failed to send id token",
			Type:    "TokenError",
		})
		return
	}

	result, err := c.LoginService.LoginGoogle(req.IdToken)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, model.APIResponse{
			Error:   true,
			Message: err.Error(),
			Type:    "LoginError",
			Data:    nil,
		})
		return
	}
	ctx.JSON(http.StatusOK, model.APIResponse{
		Error:   false,
		Message: "Login successful",
		Type:    "Success",
		Data:    result,
	})

}
