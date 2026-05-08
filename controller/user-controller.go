package controller

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/muhammadfarrasfajri/filantropi/model"
	"github.com/muhammadfarrasfajri/filantropi/service"
)

type UserController struct {
	UserService *service.UserService
}

func NewUserController(userService *service.UserService) *UserController {
	return &UserController{
		UserService: userService,
	}
}

// @Summary      Get User Profile (Donor)
// @Description  Mengambil data profil user yang sedang login (Donor) berdasarkan token JWT yang dikirimkan di header Authorization.
// @Tags         User
// @Produce      json
// @Security     BearerAuth
// @Success      200      {object}  model.APIResponse{data=model.User} "Berhasil mengambil profil user"
// @Failure      401      {object}  model.APIResponse "Unauthorized: Token tidak valid atau expired"
// @Failure      404      {object}  model.APIResponse "User tidak ditemukan di database"
// @Router       /api/user/profile/donors [get]
func (c UserController) FindById(ctx *gin.Context) {

	userID := ctx.GetString("user_id")
	if userID == "" {
		ctx.JSON(http.StatusUnauthorized, model.APIResponse{
			Error:   true,
			Message: "Unauthorized: Token tidak valid atau user_id tidak ditemukan",
			Type:    "Auth Error",
			Data:    nil,
		})
		return
	}

	// 2. Cari user ke database
	result, err := c.UserService.FindById(userID)
	if err != nil {
		// 3. Ganti 400 (BadRequest) menjadi 404 (NotFound)
		ctx.JSON(http.StatusNotFound, model.APIResponse{
			Error:   true,
			Message: err.Error(),
			Type:    "Error find user",
			Data:    nil,
		})
		return
	}

	// 4. Response Sukses
	ctx.JSON(http.StatusOK, model.APIResponse{
		Error:   false,
		Message: "Success",
		Type:    "Find User",
		Data:    result,
	})
}

// @Summary      Get Beneficiary Profile
// @Description  Mengambil profil lengkap Penerima Manfaat (Individual/Organization) berdasarkan user_id di token.
// @Tags         User
// @Produce      json
// @Security     BearerAuth
// @Success      200      {object}  model.APIResponse{data=object} "Berhasil (Data dinamis tergantung beneficiary_type)"
// @Failure      401      {object}  model.APIResponse "Unauthorized"
// @Failure      404      {object}  model.APIResponse "User/Profile tidak ditemukan"
// @Router       /api/user/profile/beneficiaries [get]
func (c UserController) FindBeneficiaryById(ctx *gin.Context) {
	userID := ctx.GetString("user_id")
	if userID == "" {
		ctx.JSON(http.StatusUnauthorized, model.APIResponse{
			Error:   true,
			Message: "Unauthorized",
			Type:    "Auth Error",
		})
		return
	}

	user, profile, err := c.UserService.FindBeneficiaryById(userID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, model.APIResponse{
			Error:   true,
			Message: err.Error(),
			Type:    "Error find user",
		})
		return
	}

	// 1. Buat Map dasar untuk data yang selalu ada (Common Fields)
	combinedData := gin.H{
		"id":               user.ID,
		"email":            user.Email,
		"wallet_address":   user.WalletAddress,
		"role":             user.Role,
		"is_verified":      user.Isverified,
		"avatar_url":       user.AvatarUrl,
		"beneficiary_type": user.BeneficiaryType,
		"full_name":        profile.FullName,
		"phone_number":     profile.PhoneNumber,
		"alamat":           profile.Alamat,
		"bio_description":  profile.BioDescription,
		"photo_profile":    profile.PhotoProfile,
		"nik":              profile.Nik,
		"url_ktp":          profile.UrlKTP,
	}

	// 2. Tambahkan field spesifik secara dinamis
	if user.BeneficiaryType == "individual" {
		combinedData["jenis_kelamin"] = profile.JenisKelamin
		combinedData["tempat_lahir"] = profile.TempatLahir
		combinedData["tanggal_lahir"] = profile.TanggalLahir
		combinedData["pekerjaan"] = profile.Pekerjaan

	} else if user.BeneficiaryType == "organization" {
		combinedData["registration_number"] = profile.RegistrationNumber
		combinedData["npwp"] = profile.Npwp
		combinedData["pic"] = profile.PIC
	}

	// 3. Response hanya akan berisi field yang relevan
	ctx.JSON(http.StatusOK, model.APIResponse{
		Error:   false,
		Message: "Success",
		Type:    "Find User",
		Data:    combinedData,
	})
}

// @Summary      Get Beneficiary Profile
// @Description  Mengambil profil lengkap Penerima Manfaat. Field yang dikembalikan dinamis: jika 'individual' akan muncul NIK & data pribadi, jika 'organization' akan muncul No. Registrasi & NPWP.
// @Tags         User
// @Produce      json
// @Security     BearerAuth
// @Success      200      {object}  model.APIResponse{data=object} "Berhasil mengambil profil (Data berisi gabungan User & Profile)"
// @Failure      401      {object}  model.APIResponse "Unauthorized"
// @Failure      404      {object}  model.APIResponse "Data tidak ditemukan"
// @Router       /api/user/profile/beneficiaries [get]
func (c UserController) UpdateDonors(ctx *gin.Context) {
	// 1. Ambil ID dari JWT Middleware (Bukan dari Body!)
	userID := ctx.GetString("user_id")
	if userID == "" {
		ctx.JSON(http.StatusUnauthorized, model.APIResponse{
			Error:   true,
			Message: "Unauthorized: ID tidak ditemukan di token",
			Type:    "AuthError",
		})
		return
	}

	var req model.UpdateDonorsRequest

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

	// 3. Bind JSON body
	if err := ctx.ShouldBind(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, model.APIResponse{
			Error:   true,
			Message: "Invalid data format",
			Type:    "ValidationError",
		})
		return
	}

	// 4. Kirim userID hasil extraksi JWT ke Service
	err = c.UserService.UpdateDonors(userID, req.WalletAddress, req.FullName, req.PhotoProfile)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, model.APIResponse{
			Error:   true,
			Message: err.Error(),
			Type:    "UpdateUserError",
		})
		return
	}

	ctx.JSON(http.StatusOK, model.APIResponse{
		Error:   false,
		Message: "Update data user success",
		Type:    "UpdateUser",
	})
}

// @Summary      Update Beneficiary Profile
// @Description  Memperbarui data profil Penerima Manfaat. Menggunakan multipart/form-data untuk mendukung upload foto profil.
// @Tags         User
// @Accept       multipart/form-data
// @Produce      json
// @Security     BearerAuth
// @Param        full_name            formData  string  false  "Nama Lengkap"
// @Param        phone_number         formData  string  false  "Nomor Telepon"
// @Param        alamat               formData  string  false  "Alamat Lengkap"
// @Param        bio_description      formData  string  false  "Deskripsi Singkat/Bio"
// @Param        nik                  formData  string  false  "NIK (Jika tipe individu)"
// @Param        jenis_kelamin        formData  string  false  "Jenis Kelamin"
// @Param        tempat_lahir         formData  string  false  "Tempat Lahir"
// @Param        tanggal_lahir        formData  string  false  "Tanggal Lahir"
// @Param        pekerjaan            formData  string  false  "Pekerjaan"
// @Param        registration_number  formData  string  false  "Nomor Registrasi (Jika tipe organisasi)"
// @Param        npwp                 formData  string  false  "NPWP (Jika tipe organisasi)"
// @Param        photo_profile        formData  file    false  "File Foto Profil (.jpg, .png)"
// @Success      200      {object}  model.APIResponse "Profil berhasil diperbarui"
// @Failure      400      {object}  model.APIResponse "Format data salah"
// @Failure      401      {object}  model.APIResponse "Unauthorized"
// @Failure      500      {object}  model.APIResponse "Gagal memproses data di server"
// @Router       /api/user/profile/update-beneficiaries [put]
func (c UserController) UpdateProfileBeneficiaries(ctx *gin.Context) {
	userID := ctx.GetString("user_id")
	if userID == "" {
		ctx.JSON(http.StatusUnauthorized, model.APIResponse{
			Error:   true,
			Message: "Unauthorized",
			Type:    "AuthError",
		})
		return
	}

	var req model.BeneficiaryProfile

	// 1. Bind data form-data terlebih dahulu (Nama, NIK, dll)
	// ShouldBind akan otomatis mendeteksi multipart form
	if err := ctx.ShouldBind(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, model.APIResponse{
			Error:   true,
			Message: err.Error(),
			Type:    "ValidationError",
		})
		return
	}

	// 2. Handle Upload File jika ada
	fileProfile, err := ctx.FormFile("photo_profile")
	if err == nil {
		// Ada file yang diupload, proses simpan!
		pathProfile := "public/uploads/profile/" + fmt.Sprintf("%d_%s", time.Now().Unix(), fileProfile.Filename)

		if errSave := ctx.SaveUploadedFile(fileProfile, pathProfile); errSave != nil {
			ctx.JSON(http.StatusInternalServerError, model.APIResponse{
				Error:   true,
				Message: "Gagal menyimpan foto",
				Type:    "ProfileError",
			})
			return
		}

		// Timpa nilai PhotoProfile dengan path file yang baru
		req.PhotoProfile = pathProfile
	}

	fileKTP, err := ctx.FormFile("url_ktp")
	if err == nil {
		// Ada file yang diupload, proses simpan!
		pathKTP := "public/uploads/profile/" + fmt.Sprintf("%d_%s", time.Now().Unix(), fileKTP.Filename)

		if errSave := ctx.SaveUploadedFile(fileKTP, pathKTP); errSave != nil {
			ctx.JSON(http.StatusInternalServerError, model.APIResponse{
				Error:   true,
				Message: "failed save KTP",
				Type:    "KTPError",
			})
			return
		}
		// Timpa nilai PhotoProfile dengan path file yang baru
		req.UrlKTP = pathKTP
	}

	// 3. Eksekusi Service
	err = c.UserService.UpdateProfileBeneficiaries(context.Background(), userID, req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, model.APIResponse{
			Error:   true,
			Message: err.Error(),
			Type:    "UpdateUserError",
		})
		return
	}

	ctx.JSON(http.StatusOK, model.APIResponse{
		Error:   false,
		Message: "Update data user success",
		Type:    "UpdateUser",
	})
}

func (c UserController) PostWallet(ctx *gin.Context) {
	wallet := ctx.Param("wallet")
	err := c.UserService.PostWallet(wallet)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, model.APIResponse{
			Error:   true,
			Message: err.Error(),
			Type:    "PostWalletError",
		})
		return
	}
	ctx.JSON(http.StatusOK, model.APIResponse{
		Error:   false,
		Message: "Wallet added successfully",
		Type:    "PostWallet",
	})
}
