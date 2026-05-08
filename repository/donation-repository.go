package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/muhammadfarrasfajri/filantropi/model"
)

type DonationRepository struct {
	DB *sql.DB
}

func NewDonationRepository(db *sql.DB) *DonationRepository {
	return &DonationRepository{
		DB: db,
	}
}

// repository/donation_repository.go
func (r *DonationRepository) CreateDonation(donation model.Donation) (model.Donation, error) {
	query := `
        INSERT INTO donations (
            id, campaign_id, user_id, wallet_address, amount, 
            message, status, transaction_hash, is_anonymous, created_at
        ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, NOW())`

	_, err := r.DB.Exec(query,
		donation.ID, donation.CampaignID, donation.UserID, donation.WalletAddress,
		donation.Amount, donation.Message, donation.Status,
		donation.TransactionHash, donation.IsAnonymous,
	)

	if err != nil {
		return donation, err
	}
	return donation, nil
}

func (r *DonationRepository) UpdateSuccess(txHash string, campaignID string, amount float64) error {
	tx, err := r.DB.Begin() // Mulai transaksi DB
	if err != nil {
		return err
	}

	// 1. Update status donasi jadi 'success'
	_, err = tx.Exec("UPDATE donations SET status = 'success' WHERE transaction_hash = ?", txHash)
	if err != nil {
		tx.Rollback()
		return err
	}

	// 2. Tambah saldo current_amount di tabel campaigns
	_, err = tx.Exec("UPDATE campaigns SET current_amount = current_amount + ? WHERE id = ?", amount, campaignID)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (r *DonationRepository) UpdateToFailed(txHash string) error {
	query := `UPDATE donations SET status = 'failed' WHERE transaction_hash = ?`

	_, err := r.DB.Exec(query, txHash)
	if err != nil {
		return err
	}

	return nil
}

func (r *DonationRepository) GetHistoryByUserID(ctx context.Context, userID string) ([]model.DonationHistoryResponse, error) {
	var history []model.DonationHistoryResponse

	query := `
        SELECT
            d.id, d.transaction_hash, d.amount, d.status, d.created_at,
            c.title as campaign_name
        FROM donations d
        JOIN campaigns c ON d.campaign_id = c.id
        WHERE d.user_id = ?
        ORDER BY d.created_at DESC`

	rows, err := r.DB.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var h model.DonationHistoryResponse
		err := rows.Scan(
			&h.ID, &h.TxHash, &h.Amount, &h.Status, &h.CreatedAt,
			&h.CampaignName,
		)
		if err != nil {
			return nil, err
		}
		history = append(history, h)
	}

	return history, nil
}

// internal/repository/donation_repository.go

func (r *DonationRepository) GetUserWallet(userID string) (string, error) {
	var wallet sql.NullString
	err := r.DB.QueryRow(`SELECT wallet_address FROM users WHERE id = ?`, userID).Scan(&wallet)
	if err != nil {
		return "", err
	}
	return wallet.String, nil
}

func (r *DonationRepository) GetCampaignByWallet(walletAddr string) (model.Campaign, error) {
	var cp model.Campaign
	query := `SELECT id, title, image_banner FROM campaigns WHERE wallet_address = ? LIMIT 1`
	err := r.DB.QueryRow(query, walletAddr).Scan(&cp.ID, &cp.Title, &cp.ImageBanner)
	return cp, err
}

// internal/repository/campaign_repository.go

func (r *DonationRepository) ProcessIncomingDonation(walletCampaign, donaturAddress string, amount float64, txHash, txTime string) error {
	var campaignID string

	// 1. Cari ID Kampanye dengan LOWER() untuk anti case-sensitive
	queryFind := "SELECT id FROM campaigns WHERE LOWER(wallet_address) = ?"
	err := r.DB.QueryRow(queryFind, walletCampaign).Scan(&campaignID)
	if err != nil {
		return fmt.Errorf("kampanye dengan dompet %s tidak ditemukan: %w", walletCampaign, err)
	}

	// 2. Mulai Transaksi SQL
	tx, err := r.DB.Begin()
	if err != nil {
		return err
	}

	// 3. Insert ke tabel donations
	queryInsert := `INSERT INTO donations (campaign_id, wallet_address_campaign, donatur_address, amount, tx_hash, created_at) VALUES (?, ?, ?, ?, ?, ?)`
	_, err = tx.Exec(queryInsert, campaignID, walletCampaign, donaturAddress, amount, txHash, txTime)

	if err != nil {
		tx.Rollback()
		// Jika error karena tx_hash sudah ada (Duplicate Entry), anggap sukses agar tidak di-retry Alchemy
		if strings.Contains(err.Error(), "Duplicate entry") {
			return nil
		}
		return fmt.Errorf("gagal mencatat riwayat donasi: %v", err)
	}

	// 4. Tambah saldo di tabel campaigns
	queryUpdate := `UPDATE campaigns SET current_amount = current_amount + ? WHERE id = ?`
	_, err = tx.Exec(queryUpdate, amount, campaignID)

	if err != nil {
		tx.Rollback()
		return fmt.Errorf("gagal update saldo kampanye: %v", err)
	}

	// 5. Simpan semua perubahan
	return tx.Commit()
}

func (r *DonationRepository) GetWalletStats(walletAddress string) (float64, []model.BeneficiaryHistory, error) {
	safeWallet := strings.ToLower(walletAddress)

	var totalBalance float64
	var histories []model.BeneficiaryHistory

	// 1. Ambil Total Saldo dari tabel campaigns
	queryBalance := "SELECT current_amount FROM campaigns WHERE LOWER(wallet_address) = ?"
	err := r.DB.QueryRow(queryBalance, safeWallet).Scan(&totalBalance)
	if err != nil {
		if err == sql.ErrNoRows {
			// Jika dompet tidak ketemu, return error agar bisa ditangani Controller
			return 0, nil, fmt.Errorf("wallet tidak ditemukan di database")
		}
		return 0, nil, err
	}

	// 2. Ambil Riwayat Donasi dari tabel donations (Diurutkan dari yang terbaru)
	queryHistory := `
		SELECT donatur_address, amount, tx_hash, created_at 
		FROM donations 
		WHERE LOWER(wallet_address_campaign) = ? 
		ORDER BY created_at DESC
	`
	rows, err := r.DB.Query(queryHistory, safeWallet)
	if err != nil {
		return 0, nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var h model.BeneficiaryHistory
		var createdAtRaw []uint8 // Menangani tipe DATETIME dari MySQL

		if err := rows.Scan(&h.DonaturAddress, &h.Amount, &h.TxHash, &createdAtRaw); err != nil {
			return 0, nil, err
		}

		h.CreatedAt = string(createdAtRaw)
		histories = append(histories, h)
	}

	// Jika histories kosong (belum ada donasi), pastikan return array kosong [], bukan null
	if histories == nil {
		histories = []model.BeneficiaryHistory{}
	}

	return totalBalance, histories, nil
}

func (r *DonationRepository) GetDonaturHistory(donaturWallet string) ([]model.DonorsHistory, error) {
	// Pastikan huruf kecil semua agar cocok dengan database
	safeDonatur := strings.ToLower(donaturWallet)
	var histories []model.DonorsHistory

	// Query simpel: Cari semua donasi di mana donatur_address = dompet user
	queryHistory := `
		SELECT wallet_address_campaign, amount, tx_hash, created_at 
		FROM donations 
		WHERE LOWER(donatur_address) = ? 
		ORDER BY created_at DESC
	`

	rows, err := r.DB.Query(queryHistory, safeDonatur)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var h model.DonorsHistory
		var createdAtRaw []uint8

		if err := rows.Scan(&h.CampaignWallet, &h.Amount, &h.TxHash, &createdAtRaw); err != nil {
			return nil, err
		}

		h.CreatedAt = string(createdAtRaw)
		histories = append(histories, h)
	}

	// Cegah balasan 'null' jika user belum pernah donasi sama sekali
	if histories == nil {
		histories = []model.DonorsHistory{}
	}

	return histories, nil
}

func (r *DonationRepository) GetTotalCollectedByCampaign(campaignWallet string) (float64, error) {
	safeWallet := strings.ToLower(campaignWallet)
	var totalBalance float64

	// Ambil langsung dari kolom current_amount yang sudah otomatis terupdate oleh Webhook
	query := "SELECT current_amount FROM campaigns WHERE LOWER(wallet_address) = ?"
	err := r.DB.QueryRow(query, safeWallet).Scan(&totalBalance)

	if err != nil {
		return 0, err
	}

	return totalBalance, nil
}
