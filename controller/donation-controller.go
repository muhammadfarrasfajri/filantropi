package controller

import (
	"fmt"
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

func (c *DonationController) CreateDonation(ctx *gin.Context) {
	var input model.DonationInput

	// 1. Validasi Input JSON
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, model.APIResponse{
			Error:   true,
			Message: "Input tidak valid",
			Type:    "ValidationError",
		})
		return
	}

	// 2. Ambil UserID dari middleware JWT
	userID := ctx.GetString("user_id")

	// 3. Panggil Service (Di sini proses simpan DB + Verifikasi Alchemy berjalan)
	donation, err := c.DonationService.CreateDonation(input, userID)

	if err != nil {
		// Jika errornya karena verifikasi blockchain (misal: hash palsu atau nominal beda)
		// Kita gunakan 422 (Unprocessable Entity) atau 400
		ctx.JSON(http.StatusUnprocessableEntity, model.APIResponse{
			Error:   true,
			Message: "Verifikasi donasi gagal: " + err.Error(),
			Type:    "BlockchainVerificationError",
		})
		return
	}

	// 4. Response Berhasil
	// Sekarang pesannya lebih akurat karena data sudah diverifikasi on-chain
	ctx.JSON(http.StatusOK, model.APIResponse{
		Error:   false,
		Message: "Donasi berhasil diverifikasi dan saldo kampanye telah diperbarui",
		Data:    donation,
	})
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

// // GET /api/wallets/:address/history
// func (c *DonationController) GetWalletHistory(ctx *gin.Context) {
// 	walletAddr := ctx.Param("address")

// 	// Validasi format wallet (0x + 40 char)
// 	if !strings.HasPrefix(walletAddr, "0x") || len(walletAddr) != 42 {
// 		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Format wallet address tidak valid"})
// 		return
// 	}

// 	history, err := c.DonationService.GetUserDonationHistory(ctx.Request.Context(), walletAddr)
// 	if err != nil {
// 		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}

// 	ctx.JSON(http.StatusOK, gin.H{
// 		"status": "success",
// 		"data":   history,
// 	})
// }

// internal/controller/donation_controller.go

func (c *DonationController) GetWalletHistory(ctx *gin.Context) {
	// 1. Ambil User ID dari Token JWT
	walletAddress := ctx.Param("wallet")

	// 2. Panggil Service
	history, err := c.DonationService.GetWalletHistory(fmt.Sprintf("%v", walletAddress))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Gagal menarik riwayat dari blockchain",
		})
		return
	}

	// 3. Cek jika data kosong
	if len(history) == 0 {
		ctx.JSON(http.StatusOK, gin.H{
			"status":  "success",
			"message": "Belum ada riwayat donasi",
			"data":    []interface{}{},
		})
		return
	}

	// 4. Kirim respon sukses
	ctx.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Riwayat donasi berhasil diambil",
		"data":    history,
	})
}

// internal/controller/donation_controller.go

// func (c *DonationController) HandleAlchemyWebhook(ctx *gin.Context) {
// 	var payload model.AlchemyWebhookPayload

// 	// 1. Tangkap JSON dari Alchemy
// 	if err := ctx.ShouldBindJSON(&payload); err != nil {
// 		fmt.Printf("[WEBHOOK ERROR] Gagal baca JSON: %v\n", err)
// 		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Format JSON tidak valid"})
// 		return
// 	}

// 	// 2. Looping aktivitas transaksi (karena 1 hash bisa punya beberapa transfer)
// 	for _, activity := range payload.Event.Activity {

// 		// Pastikan ini adalah transfer token (erc20)
// 		if activity.Category == "erc20" {

// 			// 3. LOGIKA SORTIR: Cek apakah dompet penerima (ToAddress) ada di DB Kampanye kita
// 			campaign, err := c.DonationService.GetCampaignByWallet(activity.ToAddress)

// 			if err == nil && campaign.ID != "" {
// 				// MATCH! Ini adalah donasi untuk sistem WANAMA
// 				fmt.Printf("\n[🚨 DONASI MASUK!] %f Token dari %s ke Kampanye: %s\n",
// 					activity.Value, activity.FromAddress, campaign.Title)

// 				// 4. Siapkan data untuk disimpan ke DB
// 				txData := model.TransactionData{
// 					TxHash: activity.Hash,
// 					Date:   time.Now().Format(time.RFC3339), // Waktu saat ini
// 					Type:   "Masuk",
// 					Amount: fmt.Sprintf("%.2f", activity.Value),
// 					FromTo: activity.FromAddress,
// 				}

// 				// 5. Simpan ke database menggunakan fungsi yang sudah kita buat sebelumnya
// 				errSave := c.DonationService.SaveDonation(txData, activity.ToAddress)
// 				if errSave != nil {
// 					fmt.Printf("[WEBHOOK ERROR] Gagal simpan ke DB: %v\n", errSave)
// 				} else {
// 					fmt.Printf("[WEBHOOK SUCCESS] Donasi berhasil dicatat di DB lokal!\n")
// 				}

// 			} else {
// 				// BUKAN KAMPANYE KITA. Abaikan saja.
// 				// (Ini berarti ada user di luar sana yang transfer Filantropy ke dompet pribadi temannya)
// 				fmt.Printf("[WEBHOOK IGNORE] Transfer terdeteksi, tapi bukan ke dompet kampanye.\n")
// 			}
// 		}
// 	}

// 	// 6. WAJIB: Selalu kembalikan 200 OK agar Alchemy tahu kita sudah menerimanya
// 	ctx.Status(http.StatusOK)
// }
