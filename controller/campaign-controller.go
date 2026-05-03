package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/muhammadfarrasfajri/filantropi/model"
	"github.com/muhammadfarrasfajri/filantropi/service"
)

type CampaignController struct {
	CampaignService *service.CampaignService
}

func NewCampaignController(campaignService *service.CampaignService) *CampaignController {
	return &CampaignController{
		CampaignService: campaignService,
	}
}

// @Summary      Create New Campaign
// @Description  Membuat penggalangan dana baru. Membutuhkan login sebagai Beneficiary dan upload gambar banner.
// @Tags         Campaign
// @Security     BearerAuth
// @Accept       multipart/form-data
// @Produce      json
// @Param        title                 formData  string  true  "Judul Kampanye"
// @Param        description           formData  string  true  "Deskripsi Lengkap"
// @Param        target_amount         formData  int     true  "Target Dana (Rp)"
// @Param        deadline              formData  string  true  "Tenggat Waktu (YYYY-MM-DD)"
// @Param        campaign_category_id  formData  int     true  "ID Kategori"
// @Param        image_banner          formData  file    true  "Gambar Banner Kampanye"
// @Success      201      {object}  model.APIResponse{data=object} "Kampanye Berhasil Dibuat"
// @Failure      400      {object}  model.APIResponse "Input tidak valid"
// @Router       /api/campaign [post]
func (c CampaignController) CreateCampaign(ctx *gin.Context) {
	var req model.Campaign

	// 1. Ambil UserID dari Middleware
	userID := ctx.GetString("user_id")
	if userID == "" {
		ctx.JSON(http.StatusUnauthorized, model.APIResponse{
			Error:   true,
			Message: "Unauthorized: Sesi berakhir, silakan login kembali",
			Type:    "AuthError",
		})
		return
	}
	req.UserID = userID

	// 2. Bind data teks dari Form
	// Gunakan ShouldBind agar bisa baca multipart-form
	if err := ctx.ShouldBind(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, model.APIResponse{
			Error:   true,
			Message: "Data tidak valid: " + err.Error(),
			Type:    "ValidationError",
		})
		return
	}

	// 3. PROSES UPLOAD GAMBAR
	file, err := ctx.FormFile("image_banner")
	if err != nil {
		// Jika wajib ada gambar, return error. Jika tidak, lewati saja.
		ctx.JSON(http.StatusBadRequest, model.APIResponse{
			Error:   true,
			Message: "Image banner wajib diunggah",
			Type:    "UploadError",
		})
		return
	}

	// Tentukan lokasi simpan (Contoh: public/uploads/campaigns/namafile.png)
	// Sebaiknya tambahkan userID atau timestamp agar nama file unik
	path := "public/uploads/campaigns/" + userID + "-" + file.Filename
	if err := ctx.SaveUploadedFile(file, path); err != nil {
		ctx.JSON(http.StatusInternalServerError, model.APIResponse{
			Error:   true,
			Message: "Gagal menyimpan gambar",
			Type:    "InternalError",
		})
		return
	}

	// Simpan path ke struct sebelum dikirim ke Service
	req.ImageBanner = path

	// 4. Panggil Service
	result, err := c.CampaignService.CreateCampaign(req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, model.APIResponse{
			Error:   true,
			Message: err.Error(),
			Type:    "CreateCampaignError",
		})
		return
	}

	ctx.JSON(http.StatusOK, model.APIResponse{
		Error:   false,
		Message: "Berhasil membuat penggalangan dana!",
		Type:    "Success",
		Data:    result,
	})
}

// @Summary      Get Active Campaigns
// @Description  Mengambil daftar semua kampanye yang berstatus 'active'. Digunakan oleh donatur untuk melihat daftar donasi yang tersedia.
// @Tags         Campaign
// @Produce      json
// @Success      200      {object}  model.APIResponse{data=[]model.Campaign} "Daftar kampanye aktif"
// @Failure      500      {object}  model.APIResponse "Gagal mengambil data dari database"
// @Router       /api/campaign/active [get]
func (c *CampaignController) GetActiveCampaigns(ctx *gin.Context) {
	// Kita panggil yang statusnya 'active' buat donatur
	campaigns, err := c.CampaignService.GetCampaignByStatus("active")

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, model.APIResponse{
			Error:   true,
			Message: "Gagal mengambil data kampanye: " + err.Error(),
			Type:    "DatabaseError",
		})
		return
	}

	ctx.JSON(http.StatusOK, model.APIResponse{
		Error:   false,
		Message: "Berhasil mengambil kampanye aktif",
		Data:    campaigns,
	})
}

// @Summary      Get Campaign Detail
// @Description  Mengambil detail lengkap satu kampanye berdasarkan Slug (URL unik).
// @Tags         Campaign
// @Produce      json
// @Param        slug   path      string  true  "Slug Kampanye (contoh: bantu-masjid-al-ikhlas)"
// @Success      200    {object}  model.APIResponse{data=model.Campaign} "Detail Kampanye ditemukan"
// @Failure      404    {object}  model.APIResponse "Kampanye tidak ditemukan"
// @Router       /api/campaign/{slug} [get]
func (c *CampaignController) GetCampaignDetail(ctx *gin.Context) {
	// Ambil slug dari URL: /api/campaigns/bantu-masjid-al-ikhlas
	slug := ctx.Param("slug")

	campaign, err := c.CampaignService.GetCampaignBySlug(slug)
	if err != nil {
		ctx.JSON(http.StatusNotFound, model.APIResponse{
			Error:   true,
			Message: "Kampanye tidak ditemukan",
			Type:    "NotFound",
		})
		return
	}

	ctx.JSON(http.StatusOK, model.APIResponse{
		Error:   false,
		Message: "Detail Kampanye ditemukan",
		Data:    campaign,
	})
}

// @Summary      Get H Campaigns
// @Description  Mengambil semua daftar kampanye yang dibuat oleh user (Beneficiary) yang sedang login berdasarkan token JWT.
// @Tags         Campaign
// @Produce      json
// @Security     BearerAuth
// @Success      200      {object}  model.APIResponse{data=[]model.Campaign} "Berhasil mengambil daftar kampanye milik sendiri"
// @Failure      401      {object}  model.APIResponse "Unauthorized: Silakan login kembali"
// @Failure      500      {object}  model.APIResponse "Gagal mengambil data dari database"
// @Router       /api/campaign/me [get]
func (c *CampaignController) GetMyCampaigns(ctx *gin.Context) {
	// 1. Ambil ID dari JWT yang sudah diset di Middleware
	userID := ctx.GetString("user_id")
	if userID == "" {
		ctx.JSON(http.StatusUnauthorized, model.APIResponse{
			Error:   true,
			Message: "Unauthorized: Silakan login kembali",
		})
		return
	}

	// 2. Panggil Service
	campaigns, err := c.CampaignService.GetUserCampaigns(userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, model.APIResponse{
			Error:   true,
			Message: "Gagal mengambil data kampanye: " + err.Error(),
		})
		return
	}

	// 3. Berikan Response
	ctx.JSON(http.StatusOK, model.APIResponse{
		Error:   false,
		Message: "Berhasil mengambil daftar kampanye milik Anda",
		Data:    campaigns,
	})
}

// @Summary      Update Campaign
// @Description  Memperbarui data kampanye (Judul, Deskripsi, Target, atau Banner). Hanya pemilik kampanye yang bisa melakukan update.
// @Tags         Campaign
// @Security     BearerAuth
// @Accept       multipart/form-data
// @Produce      json
// @Param        id                    path      string  true  "ID Kampanye"
// @Param        title                 formData  string  false "Judul Baru"
// @Param        description           formData  string  false "Deskripsi Baru"
// @Param        target_amount         formData  int     false "Target Dana Baru"
// @Param        deadline              formData  string  false "Deadline Baru (YYYY-MM-DD)"
// @Param        campaign_category_id  formData  int     false "ID Kategori Baru"
// @Param        image_banner          formData  file    false "File Banner Baru (Opsional)"
// @Success      200      {object}  model.APIResponse{data=object} "Berhasil memperbarui kampanye"
// @Failure      400      {object}  model.APIResponse "Input tidak valid"
// @Failure      403      {object}  model.APIResponse "Forbidden: Bukan pemilik kampanye"
// @Router       /api/campaign/{id} [put]
func (c *CampaignController) UpdateCampaign(ctx *gin.Context) {
	// 1. Ambil ID dari URL dan JWT
	slug := ctx.Param("slug")
	userID := ctx.GetString("user_id")

	var input model.Campaign
	if err := ctx.ShouldBind(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, model.APIResponse{
			Error: true, Message: err.Error(),
		})
		return
	}

	// 2. Handle Upload Gambar (Jika ada gambar baru)
	file, err := ctx.FormFile("image_banner")
	if err == nil { // Jika tidak error berarti ada file baru
		path := "public/uploads/campaigns/" + userID + "-" + file.Filename
		ctx.SaveUploadedFile(file, path)
		input.ImageBanner = path
	}

	// 3. Panggil Service
	result, err := c.CampaignService.UpdateCampaign(slug, userID, input)
	if err != nil {
		ctx.JSON(http.StatusForbidden, model.APIResponse{
			Error: true, Message: err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, model.APIResponse{
		Error: false, Message: "Berhasil memperbarui kampanye", Data: result,
	})
}

func (c *CampaignController) ApproveCampaign(ctx *gin.Context) {
	id := ctx.Param("id")

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
	err := c.CampaignService.ApproveCampaign(id, adminID)
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
