package controller

import (
	"fmt"
	"net/http"
	"time"

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

// @Summary      Register Donor with Google
// @Description  Mendaftarkan user baru sebagai donor menggunakan IdToken dari Google Auth.
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Param        request  body      model.User  true  "Kirimkan IdToken dari Google"
// @Success      201      {object}  model.APIResponse{data=object} "Register Berhasil"
// @Failure      400      {object}  model.APIResponse "Data tidak valid atau Token kosong"
// @Router       /api/auth/register/donor [post]
func (c RegisterController) RegisterGoogleUser(ctx *gin.Context) {
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

	fileProfile, err := ctx.FormFile("photo_profile")
	if err == nil {
		pathProfile := "public/uploads/profile/" + fmt.Sprintf("%d_%s", time.Now().Unix(), fileProfile.Filename)
		if errSave := ctx.SaveUploadedFile(fileProfile, pathProfile); errSave != nil {
			ctx.JSON(http.StatusInternalServerError, model.APIResponse{
				Error:   true,
				Message: "Failed to send profile photo",
				Type:    "ProfileError",
			})
			return
		}
		req.PhotoProfile = pathProfile
	} else {
		req.PhotoProfile = ""
	}

	result, err := c.RegisService.RegisterGoogle(req, model.BeneficiaryProfile{})
	if err != nil {
		ctx.JSON(http.StatusBadRequest, model.APIResponse{
			Error:   true,
			Message: err.Error(),
			Type:    "RegistrationError",
			Data:    nil,
		})
		return
	}

	ctx.JSON(http.StatusCreated, model.APIResponse{
		Error:   false,
		Message: "Register Success",
		Type:    "Register user",
		Data:    result,
	})
}

// @Summary      Register Beneficiary with Google
// @Description  Mendaftarkan user sebagai Penerima Manfaat (Beneficiary) dengan data profil lengkap menggunakan Google IdToken.
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Param        request  body      model.RegisterBeneficiaryReq  true  "Data User dan Profil Beneficiary"
// @Success      201      {object}  model.APIResponse{data=object} "Berhasil mendaftar sebagai beneficiary"
// @Failure      400      {object}  model.APIResponse "Data tidak lengkap atau format salah"
// @Router       /api/auth/register/beneficiary [post]
func (c RegisterController) RegisterGoogleBeneficiaries(ctx *gin.Context) {
	var req model.RegisterBeneficiaryReq

	// PERBAIKAN 1: Gunakan ShouldBind, BUKAN ShouldBindJSON
	err := ctx.ShouldBind(&req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, model.APIResponse{
			Error:   true,
			Message: err.Error(),
			Type:    "ValidationError",
		})
		return
	}

	if req.User.IdToken == "" {
		ctx.JSON(http.StatusBadRequest, model.APIResponse{
			Error:   true,
			Message: "Failed to send id token",
			Type:    "TokenError",
		})
		return
	}

	// Karena menggunakan multipart/form-data, FormFile sekarang bisa terbaca
	fileProfile, err := ctx.FormFile("photo_profile")
	if err == nil {
		pathProfile := "public/uploads/profile/" + fmt.Sprintf("%d_%s", time.Now().Unix(), fileProfile.Filename)
		if errSave := ctx.SaveUploadedFile(fileProfile, pathProfile); errSave != nil {
			ctx.JSON(http.StatusInternalServerError, model.APIResponse{
				Error:   true,
				Message: "Failed to send profile photo",
				Type:    "ProfileError",
			})
			return
		}
		req.Profile.PhotoProfile = pathProfile
	} else {
		req.Profile.PhotoProfile = ""
	}

	result, err := c.RegisService.RegisterGoogle(req.User, req.Profile)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, model.APIResponse{
			Error:   true,
			Message: err.Error(),
			Type:    "RegistrationError",
			Data:    nil,
		})
		return
	}

	ctx.JSON(http.StatusCreated, model.APIResponse{
		Error:   false,
		Message: "Register Success",
		Type:    "Register user",
		Data:    result,
	})
}
