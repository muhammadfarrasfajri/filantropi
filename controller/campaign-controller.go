package controller

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

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
	result, err := c.CampaignService.CreateCampaign(userID, req)
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
		Error: false, Message: "Berhasil memperbarui kampanye",
		Data: result,
	})
}

// internal/controller/campaign_controller.go

func (c *CampaignController) RequestDisbursement(ctx *gin.Context) {
	// Ambil ID Campaign dari URL Parameter (Misal: /api/campaigns/:id/disbursements)
	campaignID := ctx.Param("id")

	// Panggil Service yang sudah berisi validasi super ketat tadi
	err := c.CampaignService.CreateDisbursementRequest(campaignID)
	if err != nil {
		// Jika gagal (kena validasi pending, salah urutan, dll)
		// Kita tangkap pesan error-nya langsung dari Service
		ctx.JSON(http.StatusBadRequest, model.APIResponse{
			Error:   true,
			Message: err.Error(),
			Type:    "ErrRequestDisbursement",
		})
		return
	}

	// Jika sukses melewati semua rintangan di Service
	ctx.JSON(http.StatusCreated, model.APIResponse{
		Error:   false,
		Message: "Pengajuan pencairan dana berhasil dikirim! Silakan tunggu persetujuan Admin.",
		Type:    "Success",
	})
}

// internal/controller/campaign_controller.go

func (c *CampaignController) SubmitReport(ctx *gin.Context) {
	campaignID := ctx.Param("id")

	// 1. Tangkap Multipart Form (Untuk Form-Data)
	form, err := ctx.MultipartForm()
	if err != nil {
		ctx.JSON(http.StatusBadRequest, model.APIResponse{
			Error:   true,
			Message: "Format request tidak valid, gunakan form-data",
			Type:    "InvalidRequest",
		})
		return
	}

	// 2. Tangkap Deskripsi
	descriptions := form.Value["description"]
	if len(descriptions) == 0 || descriptions[0] == "" {
		ctx.JSON(http.StatusBadRequest, model.APIResponse{
			Error:   true,
			Message: "Deskripsi wajib diisi",
			Type:    "InvalidInput",
		})
		return
	}
	description := descriptions[0]

	// 3. Tangkap KUMPULAN FILE Gambar (Perhatikan 's' pada proof_images)
	files := form.File["proof_images"]
	if len(files) == 0 {
		ctx.JSON(http.StatusBadRequest, model.APIResponse{
			Error:   true,
			Message: "Minimal wajib mengunggah 1 foto bukti",
			Type:    "InvalidInput",
		})
		return
	}

	// Siapkan array kosong untuk menampung link-link gambar
	var uploadedURLs []string

	// 4. Looping untuk menyimpan setiap foto satu per satu
	for _, file := range files {

		// Buat nama unik dan simpan ke server lokal
		newFileName := fmt.Sprintf("%d_%s", time.Now().UnixNano(), file.Filename)
		uploadPath := fmt.Sprintf("./public/uploads/campaigns/reports/%s", newFileName)

		if err := ctx.SaveUploadedFile(file, uploadPath); err != nil {
			ctx.JSON(http.StatusInternalServerError, model.APIResponse{
				Error:   true,
				Message: "Gagal menyimpan salah satu file gambar",
				Type:    "InternalError",
			})
			return
		}

		// Tambahkan link publiknya ke dalam array
		finalURL := "public/uploads/campaigns/reports/" + newFileName
		uploadedURLs = append(uploadedURLs, finalURL)
	}

	// 5. Ubah Array berisi URL tadi menjadi format string JSON agar bisa disimpan di MySQL
	// Hasilnya akan berbentuk: `["https://..", "https://.."]`
	jsonURLs, _ := json.Marshal(uploadedURLs)

	// 6. Masukkan ke Struct
	reportInput := model.CampaignReportInput{
		CampaignID:  campaignID,
		Description: description,
		ProofURL:    string(jsonURLs), // Simpan string JSON-nya
	}

	// 7. Panggil Service
	errService := c.CampaignService.CreateReport(reportInput)
	if errService != nil {
		ctx.JSON(http.StatusBadRequest, model.APIResponse{
			Error:   true,
			Message: errService.Error(),
			Type:    "ErrSubmitReport",
		})
		return
	}

	// 8. Sukses!
	ctx.JSON(http.StatusCreated, model.APIResponse{
		Error:   false,
		Message: "Laporan dan seluruh foto berhasil diunggah!",
		Type:    "Success",
		Data: gin.H{
			"total_foto": len(uploadedURLs),
			"url_gambar": uploadedURLs,
		},
	})
}

// internal/controller/admin_controller.go

func (c *CampaignController) GetPendingDisbursements(ctx *gin.Context) {
	// Panggil Service
	pendingList, err := c.CampaignService.GetPendingDisbursements()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, model.APIResponse{
			Error:   true,
			Message: "Gagal mengambil daftar pengajuan pencairan",
			Type:    "InternalError",
		})
		return
	}

	// Jika tidak ada data sama sekali (antrean kosong)
	if len(pendingList) == 0 {
		ctx.JSON(http.StatusOK, model.APIResponse{
			Error:   false,
			Message: "Tidak ada pengajuan pencairan yang menunggu",
			Type:    "Success",
			Data:    []interface{}{},
		})
		return
	}

	// Jika ada data
	ctx.JSON(http.StatusOK, model.APIResponse{
		Error:   false,
		Message: "Berhasil mengambil antrean pencairan dana",
		Type:    "Success",
		Data:    pendingList,
	})
}

// A. Handler untuk GET Antrean Laporan
func (c *CampaignController) GetPendingReports(ctx *gin.Context) {
	pendingList, err := c.CampaignService.GetPendingReports()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, model.APIResponse{
			Error:   true,
			Message: "Gagal mengambil daftar antrean laporan",
			Type:    "InternalError",
		})
		return
	}

	if len(pendingList) == 0 {
		ctx.JSON(http.StatusOK, model.APIResponse{
			Error:   false,
			Message: "Tidak ada laporan yang menunggu direview",
			Type:    "Success",
			Data:    []interface{}{},
		})
		return
	}

	ctx.JSON(http.StatusOK, model.APIResponse{
		Error:   false,
		Message: "Berhasil mengambil antrean laporan",
		Type:    "Success",
		Data:    pendingList,
	})
}

// B. Handler untuk PUT / Approve Laporan
func (c *CampaignController) ApproveReport(ctx *gin.Context) {
	reportID := ctx.Param("id")

	err := c.CampaignService.ApproveReport(reportID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, model.APIResponse{
			Error:   true,
			Message: err.Error(),
			Type:    "ApproveError",
		})
		return
	}

	ctx.JSON(http.StatusOK, model.APIResponse{
		Error:   false,
		Message: "Laporan berhasil disetujui! Penggalang dana kini dapat mengajukan pencairan fase berikutnya.",
		Type:    "Success",
	})
}

// internal/controller/campaign_controller.go

func (c *CampaignController) GetMilestoneStatus(ctx *gin.Context) {
	campaignID := ctx.Param("id")

	statusData, err := c.CampaignService.GetCampaignMilestoneStatus(campaignID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, model.APIResponse{
			Error:   true,
			Message: err.Error(),
			Type:    "InternalError",
		})
		return
	}

	ctx.JSON(http.StatusOK, model.APIResponse{
		Error:   false,
		Message: "Berhasil mengambil status milestone kampanye",
		Type:    "Success",
		Data:    statusData,
	})
}

// internal/controller/campaign_controller.go

func (c *CampaignController) GetCampaignStepper(ctx *gin.Context) {
	campaignID := ctx.Param("id")

	stepper, err := c.CampaignService.GetCampaignStepper(campaignID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, model.APIResponse{
			Error:   true,
			Message: err.Error(),
			Type:    "ErrCampaignStepper",
		})
		return
	}

	ctx.JSON(http.StatusOK, model.APIResponse{
		Error:   false,
		Message: "get stepper success",
		Data:    stepper,
	})
}

// internal/controller/admin_controller.go

func (c *CampaignController) RejectReport(ctx *gin.Context) {
	reportID := ctx.Param("id")
	var input struct {
		RejectReason string `json:"reject_reason" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, model.APIResponse{
			Error:   true,
			Message: "Alasan penolakan wajib diisi",
			Type:    "ValidationError",
		})
		return
	}

	err := c.CampaignService.RejectReport(reportID, input.RejectReason)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, model.APIResponse{
			Error:   true,
			Message: err.Error(),
			Type:    "RejectError",
		})
		return
	}

	ctx.JSON(http.StatusOK, model.APIResponse{
		Error:   false,
		Message: "Laporan berhasil ditolak, user diminta mengunggah ulang",
		Type:    "Success",
	})
}

// internal/controller/campaign_controller.go

func (c *CampaignController) UpdateWalletAddress(ctx *gin.Context) {
	// 1. Tangkap ID dari URL (misal: /api/campaigns/:id/wallet)
	campaignID := ctx.Param("id")

	// 2. Tangkap Body JSON dari request
	var input struct {
		WalletAddress string `json:"wallet_address" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, model.APIResponse{
			Error:   true,
			Message: "Format request tidak valid atau wallet address kosong",
			Type:    "InvalidInput",
		})
		return
	}

	// 3. Panggil Service
	err := c.CampaignService.UpdateWalletAddress(campaignID, input.WalletAddress)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, model.APIResponse{
			Error:   true,
			Message: err.Error(),
			Type:    "UpdateWalletError",
		})
		return
	}

	// 4. Sukses
	ctx.JSON(http.StatusOK, model.APIResponse{
		Error:   false,
		Message: "Wallet address berhasil diperbarui!",
		Type:    "Success",
		Data: gin.H{
			"campaign_id":    campaignID,
			"wallet_address": input.WalletAddress,
		},
	})
}
