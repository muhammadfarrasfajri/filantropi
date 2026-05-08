package controller

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/muhammadfarrasfajri/filantropi/service"
	// Sesuaikan path import service kamu di sini
)

// Struct JSON yang sudah lengkap dengan Metadata (Waktu) dan RawContract (Token)
type AlchemyWebhookPayload struct {
	Event struct {
		Activity []struct {
			From     string  `json:"fromAddress"`
			To       string  `json:"toAddress"`
			Value    float64 `json:"value"`
			Hash     string  `json:"hash"`
			Metadata struct {
				BlockTimestamp string `json:"blockTimestamp"`
			} `json:"metadata"`
			RawContract struct {
				Address string `json:"address"`
			} `json:"rawContract"`
		} `json:"activity"`
	} `json:"event"`
}

type WebhookController struct {
	// Ganti dengan inisialisasi service milikmu
	DonationService *service.DonationService
}

func NewWebhookController(donationService *service.DonationService) *WebhookController {
	return &WebhookController{
		DonationService: donationService,
	}
}

func (c *WebhookController) HandleAlchemyWebhook(ctx *gin.Context) {
	var payload AlchemyWebhookPayload

	if err := ctx.ShouldBindJSON(&payload); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Format payload tidak valid"})
		return
	}

	for _, tx := range payload.Event.Activity {
		// 1. Tangkap waktu dari blockchain, beri cadangan waktu lokal jika kosong
		blockchainTime := tx.Metadata.BlockTimestamp
		if blockchainTime == "" {
			blockchainTime = time.Now().Format("2006-01-02 15:04:05")
		}

		// 2. Tangkap Contract Address dari token yang ditransfer
		tokenContractAddress := tx.RawContract.Address

		// 3. Lempar semua data ke Service
		err := c.DonationService.ProcessIncomingDonation(tx.To, tx.From, tx.Value, tx.Hash, blockchainTime, tokenContractAddress)

		if err != nil {
			fmt.Printf("[WEBHOOK] Gagal memproses tx %s ke dompet %s: %v\n", tx.Hash, tx.To, err)
		} else {
			// Print log sukses hanya jika nilainya > 0 (mengabaikan spam kosong)
			if tx.Value > 0 {
				fmt.Printf("[WEBHOOK] Sukses mencatat donasi %f dari %s\n", tx.Value, tx.From)
			}
		}
	}

	// Wajib membalas 200 OK agar antrean Alchemy selesai
	ctx.JSON(http.StatusOK, gin.H{"message": "Webhook diterima"})
}
