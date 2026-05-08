package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/muhammadfarrasfajri/filantropi/model"
	"github.com/muhammadfarrasfajri/filantropi/service"
)

type DonationController struct {
	DonationService *service.DonationService
}

func NewDonationController(donationController *service.DonationService) *DonationController {
	return &DonationController{
		DonationService: donationController,
	}
}

func (c *DonationController) MyHistory(ctx *gin.Context) {
	// Ambil ID dari token JWT yang sudah di-parse middleware
	userID := ctx.GetString("user_id")

	history, err := c.DonationService.GetUserHistory(ctx.Request.Context(), userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, model.APIResponse{
			Error:   true,
			Message: err.Error(),
			Type:    "Get History Donation",
		})
		return
	}

	ctx.JSON(http.StatusOK, model.APIResponse{
		Error:   false,
		Message: "Success fetch history",
		Data:    history,
	})
}

// func (c *DonationController) GetWalletHistory(ctx *gin.Context) {
// 	// 1. Ambil User ID dari Token JWT
// 	walletAddress := ctx.Param("wallet")

// 	// 2. Panggil Service
// 	history, err := c.DonationService.GetWalletHistory(fmt.Sprintf("%v", walletAddress))
// 	if err != nil {
// 		ctx.JSON(http.StatusInternalServerError, model.APIResponse{
// 			Error:   true,
// 			Message: err.Error(),
// 			Type:    "ErrGetHistory",
// 		})
// 		return
// 	}

// 	// 3. Cek jika data kosong
// 	if len(history) == 0 {
// 		ctx.JSON(http.StatusOK, model.APIResponse{
// 			Error:   false,
// 			Message: "Belum ada riwayat donasi",
// 			Data:    []interface{}{},
// 		})
// 		return
// 	}

// 	// 4. Kirim respon sukses
// 	ctx.JSON(http.StatusOK, model.APIResponse{
// 		Error:   false,
// 		Message: "Riwayat donasi berhasil diambil",
// 		Data:    history,
// 	})
// }
// func (c *DonationController) GetCurrentBalance(ctx *gin.Context) {
// 	// 1. Ambil Wallet Address dari Parameter URL (misal: /api/campaigns/:wallet/balance)
// 	walletAddress := ctx.Param("wallet")

// 	// 2. Panggil Service
// 	// ctx.Param sudah mengembalikan string, jadi tidak perlu fmt.Sprintf lagi
// 	balance, err := c.DonationService.GetCurrentBalance(walletAddress)
// 	if err != nil {
// 		ctx.JSON(http.StatusInternalServerError, model.APIResponse{
// 			Error:   true,
// 			Message: err.Error(),
// 			Type:    "ErrGetCurrentBalance",
// 		})
// 		return
// 	}

// 	// 3. Kirim respon sukses
// 	// Kita tidak perlu mengecek if balance == 0, karena kalau saldonya 0,
// 	// ya kita kembalikan saja angka 0 ke frontend. Itu data yang valid!
// 	ctx.JSON(http.StatusOK, model.APIResponse{
// 		Error:   false,
// 		Message: "Saldo berhasil diambil",
// 		Data: gin.H{
// 			"balance": balance,
// 		},
// 	})
// }

func (c *DonationController) GetWalletStats(ctx *gin.Context) {
	// Ambil wallet dari parameter URL, contoh: /api/wallet/0x123.../stats
	walletParam := ctx.Param("wallet")

	if walletParam == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Wallet address tidak boleh kosong"})
		return
	}

	// Panggil Service
	result, err := c.DonationService.GetWalletStats(walletParam)
	if err != nil {
		if err.Error() == "wallet tidak ditemukan di database" {
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data wallet"})
		return
	}

	// Kembalikan Response Sukses
	ctx.JSON(http.StatusOK, gin.H{
		"message": "Sukses mengambil data wallet",
		"data":    result,
	})
}

func (c *DonationController) GetDonaturHistory(ctx *gin.Context) {
	// Ambil dompet donatur dari URL (misal: /api/donatur/0x123.../history)
	donaturWallet := ctx.Param("wallet")

	if donaturWallet == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Wallet donatur tidak boleh kosong"})
		return
	}

	histories, err := c.DonationService.GetDonaturHistory(donaturWallet)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Sukses mengambil riwayat donasi pengguna",
		"data":    histories,
	})
}

func (c *DonationController) GetTotalCollectedByCampaign(ctx *gin.Context) {
	// Ambil parameter wallet kampanye dari URL
	campaignWallet := ctx.Param("wallet")

	if campaignWallet == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Wallet kampanye tidak boleh kosong"})
		return
	}

	total, err := c.DonationService.GetTotalCollectedByCampaign(campaignWallet)
	if err != nil {
		// Tangani jika dompet kampanye tidak ada di database
		if err.Error() == "sql: no rows in result set" {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Wallet kampanye tidak ditemukan"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil total donasi kampanye"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Sukses mengambil total dana terkumpul",
		"data": map[string]interface{}{
			"campaign_wallet": campaignWallet,
			"total_amount":    total,
		},
	})
}

func (c DonationController) DeleteWallet(ctx *gin.Context) {
	wallet := ctx.Param("wallet")
	err := c.DonationService.DeleteWalletAlchemy(wallet)
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
		Message: "Wallet deleted successfully",
		Type:    "DeleteWallet",
	})
}
