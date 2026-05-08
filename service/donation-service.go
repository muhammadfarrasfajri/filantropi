package service

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/muhammadfarrasfajri/filantropi/model"
	"github.com/muhammadfarrasfajri/filantropi/repository"
	"github.com/muhammadfarrasfajri/filantropi/utils"
)

type DonationService struct {
	DonationRepo repository.DonationRepo
	CampaignRepo repository.CampaignRepo
	UserRepo     repository.UserRepo
}

func NewDonationService(donationRepo repository.DonationRepo, campaignRepo repository.CampaignRepo, userRepo repository.UserRepo) *DonationService {
	return &DonationService{
		DonationRepo: donationRepo,
		CampaignRepo: campaignRepo,
		UserRepo:     userRepo,
	}
}

func (s *DonationService) GetUserHistory(ctx context.Context, userID string) ([]model.DonationHistoryResponse, error) {
	if userID == "" {
		return nil, errors.New("user ID is required")
	}
	return s.DonationRepo.GetHistoryByUserID(ctx, userID)
}

func (s *DonationService) GetCampaignByWallet(walletAddress string) (model.Campaign, error) {
	return s.DonationRepo.GetCampaignByWallet(walletAddress)
}

func (s *DonationService) ProcessIncomingDonation(walletCampaign, donaturAddress string, amount float64, txHash, txTime, tokenContract string) error {

	// Standarisasi jadi huruf kecil semua agar aman
	safeWallet := strings.ToLower(walletCampaign)
	safeTokenContract := strings.ToLower(tokenContract)

	// =======================================================
	// WAJIB DIISI: Masukkan Contract Address Token milikmu
	// =======================================================
	myToken := strings.ToLower(os.Getenv("CONTRACT_ADDRESS_FILANTROPI"))

	// 1. Filter Token: Buang jika bukan token kita
	if safeTokenContract != myToken {
		return nil // Return nil diam-diam agar Alchemy membuang antreannya
	}

	// 2. Filter Spam: Abaikan Null Address
	if safeWallet == "0x0000000000000000000000000000000000000000" {
		return nil
	}

	// 3. Validasi Angka: Abaikan jika donasi 0 atau minus
	if amount <= 0 {
		return nil
	}

	// 4. Kirim data yang sudah lolos seleksi ke Repository
	err := s.DonationRepo.ProcessIncomingDonation(safeWallet, donaturAddress, amount, txHash, txTime)

	if err != nil {
		// Jika error karena dompet tidak ketemu di DB lokal, buang diam-diam (spam)
		if strings.Contains(err.Error(), "no rows in result set") {
			return nil
		}

		// Return error asli jika masalahnya karena hal lain (misal koneksi DB putus)
		return err
	}

	return nil
}

func (s *DonationService) GetWalletStats(walletAddress string) (model.WalletStatsResponse, error) {

	// Panggil repository
	totalBalance, history, err := s.DonationRepo.GetWalletStats(walletAddress)
	if err != nil {
		return model.WalletStatsResponse{}, err
	}

	// Susun datanya sesuai struct Response
	response := model.WalletStatsResponse{
		WalletAddress: walletAddress,
		TotalBalance:  totalBalance,
		History:       history,
	}

	return response, nil
}
func (s *DonationService) GetDonaturHistory(donaturWallet string) ([]model.DonorsHistory, error) {
	return s.DonationRepo.GetDonaturHistory(donaturWallet)
}
func (s *DonationService) GetTotalCollectedByCampaign(campaignWallet string) (float64, error) {
	return s.DonationRepo.GetTotalCollectedByCampaign(campaignWallet)
}
func (s *DonationService) DeleteWalletAlchemy(wallet string) error {
	errWebhook := utils.RemoveWalletFromAlchemyWebhook(wallet)
	if errWebhook != nil {
		// Kamu bisa me-return error atau sekadar mencetak log,
		// tergantung kebijakan sistemmu (strict atau loose)
		fmt.Printf("[ERROR] Gagal menghapus wallet dari Alchemy: %v\n", errWebhook)
	}
	return nil
}

// func (s *DonationService) GetWalletHistory(walletAddress string) ([]model.TransactionData, error) {
// 	fmt.Printf("\n[INFO] === Sync History IN & OUT untuk Wallet: %s ===\n", walletAddress)

// 	// 1. Ambil Transaksi KELUAR (From User)
// 	fmt.Println("[DEBUG] Mengambil transaksi KELUAR...")
// 	outHistory, errOut := s.CallAlchemyAssetTransfers(walletAddress, "from")
// 	if errOut != nil {
// 		fmt.Printf("[ERROR] Gagal ambil transaksi KELUAR: %v\n", errOut)
// 		// Tetap lanjut, kita tangkap errornya tanpa membuat aplikasi crash
// 	}

// 	// 2. Ambil Transaksi MASUK (To User)
// 	fmt.Println("[DEBUG] Mengambil transaksi MASUK...")
// 	inHistory, errIn := s.CallAlchemyAssetTransfers(walletAddress, "to")
// 	if errIn != nil {
// 		fmt.Printf("[ERROR] Gagal ambil transaksi MASUK: %v\n", errIn)
// 	}

// 	var finalHistories []model.TransactionData

// 	// CACHE LOKAL: Mencegah N+1 Query Problem ke Database MySQL
// 	// Key: Wallet Address, Value: Campaign Title
// 	campaignCache := make(map[string]string)

// 	// 3. Proses Transaksi KELUAR
// 	for _, tx := range outHistory {
// 		campaignTitle, exists := campaignCache[tx.To]
// 		if !exists {
// 			campaign, err := s.DonationRepo.GetCampaignByWallet(tx.To)
// 			if err == nil {
// 				campaignTitle = campaign.Title
// 			} else {
// 				campaignTitle = "" // Tandai kosong agar tidak dicari lagi
// 			}
// 			campaignCache[tx.To] = campaignTitle
// 		}

// 		label := "Transfer ke Luar"
// 		if campaignTitle != "" {
// 			label = "Donasi: " + campaignTitle
// 		}

// 		finalHistories = append(finalHistories, model.TransactionData{
// 			TxHash: tx.Hash,
// 			Date:   tx.Metadata.BlockTimestamp,
// 			Type:   "Out",
// 			Amount: fmt.Sprintf("%.2f", tx.Value),
// 			FromTo: label,
// 		})
// 	}

// 	// 4. Proses Transaksi MASUK
// 	for _, tx := range inHistory {
// 		campaignTitle, exists := campaignCache[tx.From]
// 		if !exists {
// 			campaign, err := s.DonationRepo.GetCampaignByWallet(tx.From)
// 			if err == nil {
// 				campaignTitle = campaign.Title
// 			} else {
// 				campaignTitle = ""
// 			}
// 			campaignCache[tx.From] = campaignTitle
// 		}

// 		label := "Terima dari: " + tx.From
// 		if campaignTitle != "" {
// 			label = "Refund dari: " + campaignTitle
// 		}

// 		finalHistories = append(finalHistories, model.TransactionData{
// 			TxHash: tx.Hash,
// 			Date:   tx.Metadata.BlockTimestamp,
// 			Type:   "In",
// 			Amount: fmt.Sprintf("%.2f", tx.Value),
// 			FromTo: label,
// 		})
// 	}

// 	// 5. SORTING: Urutkan transaksi dari yang Terbaru ke yang Terlama
// 	sort.Slice(finalHistories, func(i, j int) bool {
// 		// String ISO 8601 bisa langsung dibandingkan menggunakan operator >
// 		return finalHistories[i].Date > finalHistories[j].Date
// 	})

// 	fmt.Printf("[INFO] Total Transaksi ditemukan setelah disatukan: %d\n", len(finalHistories))
// 	return finalHistories, nil
// }

// func (s *DonationService) CallAlchemyAssetTransfers(walletAddr string, direction string) ([]model.AlchemyTransfer, error) {
// 	// AMAN: Mengambil API Key dari .env.
// 	// Jika belum ada di .env, kita sediakan fallback sementara untuk testing.
// 	apiKey := os.Getenv("ALCHEMY_API_KEY")
// 	if apiKey == "" {
// 		apiKey = "EoACZFbhYDxPu8TGpit7u" // Pindahkan ini ke file .env saat production!
// 	}
// 	alchemyURL := fmt.Sprintf("https://polygon-mainnet.g.alchemy.com/v2/%s", apiKey)

// 	// Contract token USDT (Bisa dipindah ke .env juga)
// 	contractAddr := os.Getenv("USDT_CONTRACT_ADDRESS")
// 	if contractAddr == "" {
// 		contractAddr = "0x5feE45dd5435C6D502753b94c412Df59ad209258"
// 	}

// 	params := map[string]interface{}{
// 		"fromBlock":         "0x0",
// 		"contractAddresses": []string{contractAddr},
// 		"category":          []string{"erc20"},
// 		"withMetadata":      true,
// 		"excludeZeroValue":  true,
// 	}

// 	if direction == "from" {
// 		params["fromAddress"] = strings.ToLower(walletAddr)
// 	} else {
// 		params["toAddress"] = strings.ToLower(walletAddr)
// 	}

// 	payload := map[string]interface{}{
// 		"jsonrpc": "2.0",
// 		"id":      1,
// 		"method":  "alchemy_getAssetTransfers",
// 		"params":  []interface{}{params},
// 	}

// 	body, _ := json.Marshal(payload)
// 	resp, err := http.Post(alchemyURL, "application/json", bytes.NewBuffer(body))
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer resp.Body.Close()

// 	// Validasi tambahan jika HTTP Request gagal di level API Alchemy
// 	if resp.StatusCode != http.StatusOK {
// 		return nil, fmt.Errorf("api alchemy merespon dengan status: %d", resp.StatusCode)
// 	}

// 	var alchemyResp model.AlchemyTransferResponse
// 	if err := json.NewDecoder(resp.Body).Decode(&alchemyResp); err != nil {
// 		return nil, err
// 	}

// 	return alchemyResp.Result.Transfers, nil
// }

// func (s *DonationService) GetCurrentBalance(walletAddr string) (float64, error) {
// 	alchemyURL := "https://polygon-mainnet.g.alchemy.com/v2/EoACZFbhYDxPu8TGpit7u"
// 	contractAddr := "0x5feE45dd5435C6D502753b94c412Df59ad209258" // Alamat USDT/Token kamu

// 	// 1. Siapkan Payload JSON-RPC
// 	payload := map[string]interface{}{
// 		"jsonrpc": "2.0",
// 		"id":      1,
// 		"method":  "alchemy_getTokenBalances",
// 		"params":  []interface{}{walletAddr, []string{contractAddr}},
// 	}

// 	body, _ := json.Marshal(payload)
// 	resp, err := http.Post(alchemyURL, "application/json", bytes.NewBuffer(body))
// 	if err != nil {
// 		return 0, err
// 	}
// 	defer resp.Body.Close()

// 	// 2. Decode Response
// 	var res model.TokenBalanceResponse
// 	json.NewDecoder(resp.Body).Decode(&res)

// 	// 3. Konversi Hex ke Angka Desimal
// 	if len(res.Result.TokenBalances) > 0 {
// 		hexBalance := res.Result.TokenBalances[0].TokenBalance

// 		// Hapus awalan "0x" jika ada
// 		hexBalance = strings.TrimPrefix(hexBalance, "0x")

// 		// Konversi Hex ke Big Int
// 		bi := new(big.Int)
// 		bi.SetString(hexBalance, 16)

// 		// Konversi ke Float dan bagi dengan 10^18 (asumsi desimal USDT/token adalah 18)
// 		// Jika USDT di Polygon biasanya 6 desimal, ganti 1e18 jadi 1e6
// 		fbalance := new(big.Float).SetInt(bi)
// 		ethValue := new(big.Float).Quo(fbalance, big.NewFloat(1e18))

// 		result, _ := ethValue.Float64()
// 		return result, nil
// 	}

// 	return 0, nil
// }
