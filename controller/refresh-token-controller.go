package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/muhammadfarrasfajri/filantropi/model"
	"github.com/muhammadfarrasfajri/filantropi/service"
)

type RefreshTokenController struct {
	RefreshTokenService *service.RefreshTokenService
}

func NewRefreshTokenController(refreshTokenService *service.RefreshTokenService) *RefreshTokenController {
	return &RefreshTokenController{
		RefreshTokenService: refreshTokenService,
	}
}

// @Summary      Refresh Access Token
// @Description  Mendapatkan Access Token baru menggunakan Refresh Token yang valid.
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Param        request  body      model.RefreshTokenReq  true  "Kirimkan Refresh Token"
// @Success      200      {object}  model.APIResponse{data=object} "Token Berhasil Diperbarui"
// @Failure      400      {object}  model.APIResponse "Format data salah"
// @Failure      401      {object}  model.APIResponse "Refresh Token tidak valid atau expired"
// @Router       /api/auth/refresh-token [post]
func (c RefreshTokenController) RefreshToken(ctx *gin.Context) {
	var req model.RefreshTokenReq
	err := ctx.ShouldBind(&req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, model.APIResponse{
			Error:   true,
			Message: "Invalid data format",
			Type:    "ValidationError",
		})
		return
	}

	if req.RefreshToken == "" {
		ctx.JSON(http.StatusBadRequest, model.APIResponse{
			Error:   true,
			Message: "Failed to send id token",
			Type:    "TokenError",
		})
		return
	}

	result, err := c.RefreshTokenService.RefreshToken(req.RefreshToken)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, model.APIResponse{
			Error:   true,
			Message: err.Error(),
			Type:    "Error refresh token",
			Data:    nil,
		})
		return
	}

	ctx.JSON(http.StatusOK, model.APIResponse{
		Error:   false,
		Message: "Register Success",
		Type:    "Success",
		Data:    result,
	})

}

// @Summary      Logout User
// @Description  Mengeluarkan user dari sesi dengan membatalkan (invalidate) Refresh Token yang dikirim di body.
// @Tags         Authentication
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request  body      model.RefreshTokenReq  true  "Kirimkan Refresh Token yang ingin di-logout"
// @Success      200      {object}  model.APIResponse "Logout Berhasil"
// @Failure      400      {object}  model.APIResponse "Format data salah"
// @Failure      500      {object}  model.APIResponse "Gagal memproses logout di server"
// @Router       /api/auth/logout [post]
func (c RefreshTokenController) Logout(ctx *gin.Context) {

	var req model.RefreshTokenReq
	err := ctx.ShouldBindJSON(&req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, model.APIResponse{
			Error:   true,
			Message: "Invalid data format",
			Type:    "ValidationError",
		})
		return
	}

	if req.RefreshToken == "" {
		ctx.JSON(http.StatusBadRequest, model.APIResponse{
			Error:   true,
			Message: "Failed to send id token",
			Type:    "TokenError",
		})
		return
	}

	err = c.RefreshTokenService.Logout(req.RefreshToken)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, model.APIResponse{
			Error:   true,
			Message: err.Error(),
			Type:    "Error Logout",
		})
		return
	}

	ctx.JSON(http.StatusOK, model.APIResponse{
		Error:   false,
		Message: "Logout Success",
		Type:    "Success",
	})

}
