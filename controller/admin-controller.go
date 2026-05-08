package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/muhammadfarrasfajri/filantropi/model"
	"github.com/muhammadfarrasfajri/filantropi/service"
)

type AdminController struct {
	AdminService *service.AdminService
}

func NewAdminController(adminService *service.AdminService) *AdminController {
	return &AdminController{
		AdminService: adminService,
	}
}

func (c *AdminController) GetAllCampaign(ctx *gin.Context) {
	status := ctx.Query("status")
	search := ctx.Query("search")

	campaigns, err := c.AdminService.GetAllCampaign(status, search)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, model.APIResponse{
			Error:   true,
			Message: err.Error(),
			Type:    "GETAllCampaign",
		})
		return
	}
	ctx.JSON(http.StatusOK, model.APIResponse{
		Error:   false,
		Message: "Get all campaign success",
		Type:    "Success",
		Data:    campaigns,
	})
}

func (c *AdminController) GetAllUsersForAdmin(ctx *gin.Context) {
	search := ctx.Query("search")

	users, err := c.AdminService.GetAllUsersForAdmin(search)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, model.APIResponse{
			Error:   true,
			Message: err.Error(),
			Type:    "GetAllUser",
		})
		return
	}
	ctx.JSON(http.StatusOK, model.APIResponse{
		Error:   false,
		Message: "Get all user success",
		Type:    "Success",
		Data:    users,
	})
}

func (c *AdminController) GetUserDetailForAdmin(ctx *gin.Context) {
	userID := ctx.Param("id")

	users, err := c.AdminService.GetUserDetailForAdmin(userID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, model.APIResponse{
			Error:   true,
			Message: err.Error(),
			Type:    "GetUserDetail",
		})
		return
	}
	ctx.JSON(http.StatusOK, model.APIResponse{
		Error:   false,
		Message: "Get detail user success",
		Type:    "Success",
		Data:    users,
	})
}

func (c *AdminController) GetDashboardSummary(ctx *gin.Context) {
	dashboard, err := c.AdminService.GetDashboardSummary()
	if err != nil {
		ctx.JSON(http.StatusBadRequest, model.APIResponse{
			Error:   true,
			Message: err.Error(),
			Type:    "GetUserAmount",
		})
		return
	}
	ctx.JSON(http.StatusOK, model.APIResponse{
		Error:   false,
		Message: "Get all user success",
		Type:    "Success",
		Data:    dashboard,
	})
}

func (c *AdminController) ApproveCampaign(ctx *gin.Context) {
	id := ctx.Param("slug")

	// Ambil Admin ID dari JWT
	adminID := ctx.GetString("user_id")
	if adminID == "" {
		ctx.JSON(http.StatusUnauthorized, model.APIResponse{
			Error:   true,
			Message: "Unauthorized: Sesi berakhir, silakan login kembali",
			Type:    "AuthError",
		})
		return
	}

	// Panggil Service
	err := c.AdminService.ApproveCampaign(id, adminID)
	if err != nil {
		// Gunakan StatusBadRequest (400) karena ini biasanya kesalahan logika bisnis
		// atau data yang diminta tidak tersedia.
		ctx.JSON(http.StatusBadRequest, model.APIResponse{
			Error:   true,
			Message: err.Error(),
			Type:    "ApproveError",
		})
		return
	}

	ctx.JSON(http.StatusOK, model.APIResponse{
		Error:   false,
		Message: "Berhasil menyetujui kampanye. Status sekarang: ACTIVE",
		Type:    "Success",
	})
}
func (c *AdminController) RejectedCampaign(ctx *gin.Context) {
	id := ctx.Param("slug")
	type InputReject struct {
		RejectReason string `json:"reject_reason"`
	}

	var input InputReject

	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, model.APIResponse{
			Error:   true,
			Message: "invalid format data",
			Type:    "InvalidInput",
		})
		return
	}
	// Ambil Admin ID dari JWT
	adminID := ctx.GetString("user_id")
	if adminID == "" {
		ctx.JSON(http.StatusUnauthorized, model.APIResponse{
			Error:   true,
			Message: "Unauthorized: Sesi berakhir, silakan login kembali",
			Type:    "AuthError",
		})
		return
	}

	// Panggil Service
	err := c.AdminService.RejectedCampaign(id, adminID, input.RejectReason)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, model.APIResponse{
			Error:   true,
			Message: err.Error(),
			Type:    "RejectedError",
		})
		return
	}

	ctx.JSON(http.StatusOK, model.APIResponse{
		Error:   false,
		Message: "Berhasil menolak kampanye. Status sekarang: Rejected",
		Type:    "Success",
	})
}

// internal/controller/admin_controller.go

func (c *AdminController) VerifyUser(ctx *gin.Context) {
	userID := ctx.Param("id")

	// Tambahkan field 'reason' (opsional)
	var input model.InputEmaiVerified

	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, model.APIResponse{
			Error:   true,
			Message: "invalid format data",
			Type:    "InvalidInput",
		})
		return
	}

	pesanStatus, err := c.AdminService.UpdateUserVerification(userID, input)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, model.APIResponse{
			Error:   true,
			Message: err.Error(),
			Type:    "UpdateVerification",
		})
		return
	}

	ctx.JSON(http.StatusOK, model.APIResponse{
		Error:   false,
		Message: "Success",
		Data:    pesanStatus,
	})
}

func (c *AdminController) ApproveDisbursement(ctx *gin.Context) {
	// Ambil ID Disbursement dari parameter URL
	// Contoh hit URL: PUT /api/admin/disbursements/123e4567-e89b-12d3.../approve
	disbursementID := ctx.Param("id")

	// Panggil otak aplikasi (Service)
	err := c.AdminService.ApproveDisbursement(disbursementID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, model.APIResponse{
			Error:   true,
			Message: err.Error(),
			Type:    "ErrApproveDisbursement",
		})
		return
	}

	// Jika sukses, balikan JSON
	ctx.JSON(http.StatusOK, model.APIResponse{
		Error:   false,
		Message: "Pencairan dana berhasil disetujui! User sekarang dapat mulai mengunggah laporan.",
		Type:    "Success",
	})
}
