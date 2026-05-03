package repository

import (
	"context"
	"database/sql"

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
