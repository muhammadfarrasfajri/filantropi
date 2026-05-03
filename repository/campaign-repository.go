package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/muhammadfarrasfajri/filantropi/model"
)

type CampaignRepository struct {
	DB *sql.DB
}

func NewCampaignRepository(db *sql.DB) *CampaignRepository {
	return &CampaignRepository{
		DB: db,
	}
}

func (r *CampaignRepository) CreateCampaign(campaign model.Campaign) (model.Campaign, error) {
	query := `
        INSERT INTO campaigns (
            id, user_id, wallet_address, category_id, title, slug, description, story, 
            target_amount, image_banner, end_date, status
        ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := r.DB.Exec(query,
		campaign.ID, campaign.UserID, campaign.WalletAddress, campaign.CategoryID, campaign.Title,
		campaign.Slug, campaign.Description, campaign.Story,
		campaign.TargetAmount, campaign.ImageBanner, campaign.EndDate, campaign.Status,
	)

	if err != nil {
		return campaign, err
	}

	return campaign, nil
}

func (r *CampaignRepository) GetCampaignByStatus(status string) ([]model.Campaign, error) {
	query := `
    SELECT 
        c.id, 
        c.user_id,
		c.wallet_address, 
        COALESCE(bp.full_name, 'No Name') as full_name, 
        c.category_id, 
        c.title, 
        c.slug, 
        c.description, 
        c.target_amount, 
        c.current_amount, 
        c.image_banner, 
        c.status, 
        c.end_date, 
        c.created_at 
    FROM campaigns c
    LEFT JOIN beneficiary_profiles bp ON c.user_id = bp.user_id -- Kuncinya di sini: user_id = user_id
    WHERE c.status = ? 
    ORDER BY c.created_at DESC`

	rows, err := r.DB.Query(query, status)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var campaigns []model.Campaign = []model.Campaign{} // Inisialisasi slice kosong, bukan nil
	for rows.Next() {
		var c model.Campaign
		err := rows.Scan(
			&c.ID, &c.UserID, &c.WalletAddress, &c.FullName, &c.CategoryID, &c.Title, &c.Slug, &c.Description,
			&c.TargetAmount, &c.CurrentAmount, &c.ImageBanner, &c.Status,
			&c.EndDate, &c.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		campaigns = append(campaigns, c)
	}

	return campaigns, nil
}

func (r *CampaignRepository) GetBySlug(slug string) (model.Campaign, error) {
	query := `
    SELECT 
        c.id, c.user_id, c.wallet_address, 
        COALESCE(bp.full_name, 'No Name') as full_name, 
        c.category_id, c.title, c.slug, c.description, c.story,
        c.target_amount, c.current_amount, c.image_banner, c.status, 
        c.end_date, 
        c.created_at 
    FROM campaigns c
    LEFT JOIN beneficiary_profiles bp ON c.user_id = bp.user_id 
    WHERE c.slug = ? 
    LIMIT 1`

	var c model.Campaign
	err := r.DB.QueryRow(query, slug).Scan(
		&c.ID, &c.UserID, &c.WalletAddress, &c.FullName, &c.CategoryID, &c.Title, &c.Slug, &c.Description, &c.Story,
		&c.TargetAmount, &c.CurrentAmount, &c.ImageBanner, &c.Status,
		&c.EndDate, &c.CreatedAt,
	)

	if err != nil {
		return c, err
	}

	return c, nil
}

func (r *CampaignRepository) GetByUserID(userID string) ([]model.Campaign, error) {
	query := `
        SELECT id, user_id, wallet_address, category_id, title, slug, description, 
               target_amount, current_amount, image_banner, status, 
               COALESCE(end_date, '') as end_date, created_at 
        FROM campaigns 
        WHERE user_id = ? 
        ORDER BY created_at DESC`

	rows, err := r.DB.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var campaigns []model.Campaign
	for rows.Next() {
		var c model.Campaign
		err := rows.Scan(
			&c.ID, &c.UserID, &c.WalletAddress, &c.CategoryID, &c.Title, &c.Slug, &c.Description,
			&c.TargetAmount, &c.CurrentAmount, &c.ImageBanner, &c.Status,
			&c.EndDate, &c.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		campaigns = append(campaigns, c)
	}

	return campaigns, nil
}

// repository/campaign_repository.go
func (r *CampaignRepository) GetByID(id string) (model.Campaign, error) {
	var c model.Campaign
	query := `
        SELECT id, user_id, wallet_address, category_id, title, slug, description, story, 
               target_amount, current_amount, image_banner, status, created_at 
        FROM campaigns 
        WHERE id = ? LIMIT 1`

	err := r.DB.QueryRow(query, id).Scan(
		&c.ID, &c.UserID, &c.WalletAddress, &c.CategoryID, &c.Title, &c.Slug, &c.Description, &c.Story,
		&c.TargetAmount, &c.CurrentAmount, &c.ImageBanner, &c.Status, &c.CreatedAt,
	)

	return c, err
}

func (r *CampaignRepository) UpdateCampaign(campaign model.Campaign) (model.Campaign, error) {
	var setClauses []string
	var args []interface{}

	// Cek masing-masing field, jika tidak kosong/nol, masukkan ke slice
	if campaign.CategoryID != 0 {
		setClauses = append(setClauses, "category_id = ?")
		args = append(args, campaign.CategoryID)
	}
	if campaign.WalletAddress != "" {
		setClauses = append(setClauses, "wallet_address = ?")
		args = append(args, campaign.WalletAddress)
	}
	if campaign.Title != "" {
		setClauses = append(setClauses, "title = ?")
		args = append(args, campaign.Title)
	}
	if campaign.Description != "" {
		setClauses = append(setClauses, "description = ?")
		args = append(args, campaign.Description)
	}
	if campaign.Story != "" {
		setClauses = append(setClauses, "story = ?")
		args = append(args, campaign.Story)
	}
	if campaign.TargetAmount != 0 {
		setClauses = append(setClauses, "target_amount = ?")
		args = append(args, campaign.TargetAmount)
	}
	if campaign.ImageBanner != "" {
		setClauses = append(setClauses, "image_banner = ?")
		args = append(args, campaign.ImageBanner)
	}
	// Catatan: Pengecekan ImageBanner yang duplikat sudah dihapus

	// Untuk end_date, pastikan dicek sesuai tipe datanya (misal time.Time)
	if campaign.EndDate != nil {
		setClauses = append(setClauses, "end_date = ?")
		args = append(args, *campaign.EndDate)
	}

	// TAMPILKAN PENAMBAHAN STATUS DI SINI
	if campaign.Status != "" {
		setClauses = append(setClauses, "status = ?")
		args = append(args, campaign.Status)

		// Lakukan pengecekan: Jika status diubah menjadi pending, paksa reset field approval
		if campaign.Status == "pending" {
			setClauses = append(setClauses, "approved_by = ?")
			args = append(args, "") // Di-set string kosong

			setClauses = append(setClauses, "approved_at = ?")
			args = append(args, nil) // Menggunakan nil agar diubah menjadi NULL di database
		}
	}

	// Jika tidak ada data yang diupdate sama sekali
	if len(setClauses) == 0 {
		return campaign, nil
	}

	// Gabungkan semua clause dan tambahkan parameter WHERE di akhir
	query := fmt.Sprintf("UPDATE campaigns SET %s WHERE slug = ?", strings.Join(setClauses, ", "))
	args = append(args, campaign.Slug)

	_, err := r.DB.Exec(query, args...)
	if err != nil {
		return campaign, err
	}

	return campaign, nil
}

func (r *CampaignRepository) ApproveCampaign(campaignID string, adminID string) error {
	query := `
        UPDATE campaigns 
        SET status = 'active', 
            approved_by = ?, 
            approved_at = NOW() 
        WHERE id = ?`

	result, err := r.DB.Exec(query, adminID, campaignID)
	if err != nil {
		return err
	}

	// CEK: Apakah ada baris yang beneran di-update?
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return errors.New("Approve failed: campaign ID not found")
	}

	return nil
}

// repository/campaign_repo.go
func (r *CampaignRepository) GetCampaignByWallet(ctx context.Context, walletAddr string) (model.Campaign, error) {
	var cp model.Campaign
	// Kita ambil Title dan Image untuk mempercantik tampilan history di Frontend
	query := `SELECT id, title, image_banner FROM campaigns WHERE wallet_address = ? LIMIT 1`

	err := r.DB.QueryRowContext(ctx, query, walletAddr).Scan(&cp.ID, &cp.Title, &cp.ImageBanner)
	if err != nil {
		return cp, err // Akan mengembalikan sql.ErrNoRows jika tidak ketemu
	}
	return cp, nil
}
