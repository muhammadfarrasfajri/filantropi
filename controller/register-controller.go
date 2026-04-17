package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/muhammadfarrasfajri/filantropi/model"
	"github.com/muhammadfarrasfajri/filantropi/service"
)

type RegisterController struct {
	RegisService *service.RegisterService
}

func NewRegisterController(regisService *service.RegisterService) *RegisterController {
	return &RegisterController{
		RegisService: regisService,
	}
}

func (c RegisterController) RegisterGoogle(ctx *gin.Context) {
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

	result, err := c.RegisService.RegisterGoogle(req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, model.APIResponse{
			Error:   true,
			Message: err.Error(),
			Type:    "RegistrationError",
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
