package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/muhammadfarrasfajri/filantropi/model"
	"github.com/muhammadfarrasfajri/filantropi/repository"
)

type DonationService struct {
	DonationRepo repository.DonationRepo
	CampaignRepo repository.CampaighRepo
	UserRepo     repository.UserRepo
}

func NewDonationService(donationRepo repository.DonationRepo, campaignRepo repository.CampaighRepo, userRepo repository.UserRepo) *DonationService {
	return &DonationService{
		DonationRepo: donationRepo,
		CampaignRepo: campaignRepo,
		UserRepo:     userRepo,
	}
}

func (s *DonationService) CreateDonation(donationInput model.DonationInput, userId string) (model.Donation, error) {
	// 1. Simpan data awal sebagai 'pending'
	// Kita tetap pakai amount dari input user dulu buat catatan sementara
	donation := model.Donation{
		ID:              uuid.New().String(),
		CampaignID:      donationInput.CampaignID,
		UserID:          userId,
		Message:         donationInput.Message,
		Status:          "pending",
		TransactionHash: donationInput.TransactionHash,
		IsAnonymous:     donationInput.IsAnonymous,
	}

	_, err := s.DonationRepo.CreateDonation(donation)
	if err != nil {
		return donation, err
	}

	// 2. Ambil data campaign untuk mendapatkan wallet_address penerima (Manfaat)
	campaign, err := s.CampaignRepo.GetByID(donation.CampaignID)
	if err != nil {
		return donation, err
	}

	// 3. VERIFIKASI KE BLOCKCHAIN
	// Kita panggil fungsi yang bedah data 'input' ERC-20 tadi
	// Kita tidak butuh 'expectedAmount' lagi karena kita ambil langsung dari chain
	isValid, err := VerifyDonationOnChain(donation.TransactionHash, campaign.WalletAddress, donation.Amount)

	if err != nil || isValid {
		// Jika gagal (hash palsu, token salah, atau status failed), update DB jadi 'failed'
		s.DonationRepo.UpdateToFailed(donation.TransactionHash)
		return donation, fmt.Errorf("verifikasi blockchain gagal: %v", err)
	}

	// 4. UPDATE KE SUKSES
	// PENTING: Gunakan 'actualAmount' (dari Blockchain), BUKAN 'donationInput.Amount'
	err = s.DonationRepo.UpdateSuccess(donation.TransactionHash, donation.CampaignID, donation.Amount)
	if err != nil {
		return donation, err
	}

	// Update object di memory sebelum di-return ke controller
	donation.Amount = donation.Amount
	donation.Status = "success"

	return donation, nil
}

// func (s *DonationService) GetUserDonationHistory(ctx context.Context, userID string) ([]model.TransactionData, error) {
// 	// 1. Ambil Wallet User dari Database
// 	user, err := s.UserRepo.FindUserById(userID)
// 	if err != nil || user.WalletAddress == "" {
// 		return []model.TransactionData{}, nil
// 	}

// 	// 2. Request ke Alchemy (getAssetTransfers)
// 	payload := map[string]interface{}{
// 		"jsonrpc": "2.0",
// 		"id":      1,
// 		"method":  "alchemy_getAssetTransfers",
// 		"params": []map[string]interface{}{
// 			{
// 				"fromBlock":         "0x0",
// 				"fromAddress":       user.WalletAddress, // Mencari uang KELUAR
// 				"contractAddresses": []string{os.Getenv("0x5feE45dd5435C6D502753b94c412Df59ad209258")},
// 				"category":          []string{"erc20"},
// 				"withMetadata":      true,
// 			},
// 		},
// 	}

// 	jsonData, _ := json.Marshal(payload)
// 	resp, err := http.Post(os.Getenv("ALCHEMY_URL"), "application/json", bytes.NewBuffer(jsonData))
// 	if err != nil {
// 		return nil, fmt.Errorf("gagal koneksi ke blockchain: %v", err)
// 	}
// 	defer resp.Body.Close()

// 	var alchemyResp model.AlchemyTransferResponse
// 	if err := json.NewDecoder(resp.Body).Decode(&alchemyResp); err != nil {
// 		return nil, err
// 	}

// 	// 3. Mapping data Blockchain ke struct TransactionData (Hybrid dengan DB)
// 	var history []model.TransactionData
// 	for _, tx := range alchemyResp.Result.Transfers {

// 		// Cek apakah alamat 'To' adalah salah satu campaign kita
// 		campaign, err := s.CampaignRepo.GetCampaignByWallet(ctx, tx.To)

// 		var campaignName string
// 		if err != nil {
// 			campaignName = "Transfer Luar / Wallet Lain"
// 		} else {
// 			campaignName = campaign.Title
// 		}

// 		// Masukkan ke struct TransactionData sesuai punya kamu
// 		history = append(history, model.TransactionData{
// 			TxHash:  tx.Hash,
// 			Tanggal: tx.Metadata.BlockTimestamp,
// 			Tipe:    "Keluar",                      // Karena kita cari dari 'fromAddress' user
// 			Jumlah:  fmt.Sprintf("%.2f", tx.Value), // Format ke string 2 desimal
// 			DariKe:  campaignName,
// 		})
// 	}

// 	return history, nil
// }

// internal/service/donation_service.go

func (s *DonationService) GetUserHistory(ctx context.Context, userID string) ([]model.DonationHistoryResponse, error) {
	if userID == "" {
		return nil, errors.New("user ID is required")
	}
	return s.DonationRepo.GetHistoryByUserID(ctx, userID)
}

func (s *DonationService) GetCampaignByWallet(walletAddress string) (model.Campaign, error) {
	return s.DonationRepo.GetCampaignByWallet(walletAddress)
}

func (s *DonationService) GetWalletHistory(walletAddres string) ([]model.TransactionData, error) {
	fmt.Printf("\n[INFO] === Sync History IN & OUT untuk Wallet: %s ===\n", walletAddres)

	// 1. Ambil Transaksi KELUAR (From User)
	fmt.Println("[DEBUG] Mengambil transaksi KELUAR...")
	outHistory, _ := s.CallAlchemyAssetTransfers(walletAddres, "from")

	// 2. Ambil Transaksi MASUK (To User)
	fmt.Println("[DEBUG] Mengambil transaksi MASUK...")
	inHistory, _ := s.CallAlchemyAssetTransfers(walletAddres, "to")

	var finalHistories []model.TransactionData

	// 3. Proses Transaksi KELUAR
	for _, tx := range outHistory {
		campaign, err := s.DonationRepo.GetCampaignByWallet(tx.To)
		label := "Transfer ke Luar"
		if err == nil {
			label = "Donasi: " + campaign.Title
		}

		finalHistories = append(finalHistories, model.TransactionData{
			TxHash: tx.Hash,
			Date:   tx.Metadata.BlockTimestamp,
			Type:   "Out",
			Amount: fmt.Sprintf("%.2f", tx.Value),
			FromTo: label,
		})
	}

	// 4. Proses Transaksi MASUK
	for _, tx := range inHistory {
		// Cek apakah pengirimnya adalah salah satu campaign (misal: refund)
		campaign, err := s.DonationRepo.GetCampaignByWallet(tx.From)
		label := "Terima dari: " + tx.From
		if err == nil {
			label = "Refund dari: " + campaign.Title
		}

		finalHistories = append(finalHistories, model.TransactionData{
			TxHash: tx.Hash,
			Date:   tx.Metadata.BlockTimestamp,
			Type:   "In",
			Amount: fmt.Sprintf("%.2f", tx.Value),
			FromTo: label,
		})
	}

	fmt.Printf("[INFO] Total Transaksi ditemukan: %d\n", len(finalHistories))
	return finalHistories, nil
}

// internal/service/donation_service.go
func (s *DonationService) CallAlchemyAssetTransfers(walletAddr string, direction string) ([]model.AlchemyTransfer, error) {
	alchemyURL := "https://polygon-mainnet.g.alchemy.com/v2/EoACZFbhYDxPu8TGpit7u"
	contractAddr := "0x5feE45dd5435C6D502753b94c412Df59ad209258"

	params := map[string]interface{}{
		"fromBlock":         "0x0",
		"contractAddresses": []string{contractAddr},
		"category":          []string{"erc20"},
		"withMetadata":      true,
		"excludeZeroValue":  true,
	}

	// Tentukan arah transaksi
	if direction == "from" {
		params["fromAddress"] = strings.ToLower(walletAddr)
	} else {
		params["toAddress"] = strings.ToLower(walletAddr)
	}

	payload := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "alchemy_getAssetTransfers",
		"params":  []interface{}{params},
	}

	body, _ := json.Marshal(payload)
	resp, err := http.Post(alchemyURL, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var alchemyResp model.AlchemyTransferResponse
	json.NewDecoder(resp.Body).Decode(&alchemyResp)

	return alchemyResp.Result.Transfers, nil
}
