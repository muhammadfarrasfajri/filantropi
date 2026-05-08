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
        COALESCE(bp.beneficiary_type, 'Umum') as beneficiary_type, -- Ini tambahan barunya
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
		&c.ID, &c.UserID, &c.WalletAddress, &c.FullName, &c.BeneficiaryType, &c.CategoryID, &c.Title, &c.Slug, &c.Description, &c.Story,
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
			   COALESCE(reject_reason, '') as reject_reason,
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
			&c.TargetAmount, &c.CurrentAmount, &c.ImageBanner, &c.Status, &c.RejectReason,
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
               target_amount, current_amount, image_banner, status, reject_reason, created_at 
        FROM campaigns 
        WHERE id = ? LIMIT 1`

	err := r.DB.QueryRow(query, id).Scan(
		&c.ID, &c.UserID, &c.WalletAddress, &c.CategoryID, &c.Title, &c.Slug, &c.Description, &c.Story,
		&c.TargetAmount, &c.CurrentAmount, &c.ImageBanner, &c.Status, &c.RejectReason, &c.CreatedAt,
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
			setClauses = append(setClauses, "reviewed_by = ?")
			args = append(args, "") // Di-set string kosong

			setClauses = append(setClauses, "reviewed_at = ?")
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

// internal/repository/campaign_repository.go

func (r *CampaignRepository) UpdateWalletAddress(campaignID string, newWalletAddress string) error {
	query := `
		UPDATE campaigns 
		SET wallet_address = ?, updated_at = NOW() 
		WHERE id = ?
	`

	// Gunakan Exec karena kita tidak mengharapkan data kembalian (hanya mengubah data)
	result, err := r.DB.Exec(query, newWalletAddress, campaignID)
	if err != nil {
		return err
	}

	// (Opsional) Cek apakah ID kampanyenya benar-benar ada
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		// Mengembalikan error jika tidak ada baris yang berubah (mungkin ID salah)
		return errors.New("kampanye tidak ditemukan atau wallet address sudah sama")
	}

	return nil
}

func (r *CampaignRepository) CountApprovedDisbursements(campaignID string) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM campaign_disbursements WHERE campaign_id = ? AND status = 'APPROVED'`
	err := r.DB.QueryRow(query, campaignID).Scan(&count)
	return count, err
}

func (r *CampaignRepository) CountApprovedReports(campaignID string) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM campaign_reports WHERE campaign_id = ? AND status = 'APPROVED'`
	err := r.DB.QueryRow(query, campaignID).Scan(&count)
	return count, err
}

// FUNGSI A: Untuk mengecek apakah ada PENCAIRAN UANG yang masih gantung
func (r *CampaignRepository) HasPendingDisbursement(campaignID string) (bool, error) {
	var exists int
	query := `SELECT 1 FROM campaign_disbursements WHERE campaign_id = ? AND status = 'PENDING' LIMIT 1`
	err := r.DB.QueryRow(query, campaignID).Scan(&exists)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil // Tidak ada pencairan yang pending (Aman)
		}
		return false, err
	}
	return true, nil // Ada pencairan yang pending
}

// FUNGSI B: Untuk mengecek apakah ada LAPORAN NOTA yang masih gantung
func (r *CampaignRepository) HasPendingReport(campaignID string) (bool, error) {
	var exists int
	query := `SELECT 1 FROM campaign_reports WHERE campaign_id = ? AND status = 'PENDING' LIMIT 1`
	err := r.DB.QueryRow(query, campaignID).Scan(&exists)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (r *CampaignRepository) IsReportApproved(campaignID string, phase int) (bool, error) {
	var exists int
	query := `SELECT 1 FROM campaign_reports WHERE campaign_id = ? AND phase = ? AND status = 'APPROVED' LIMIT 1`
	err := r.DB.QueryRow(query, campaignID, phase).Scan(&exists)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil // Laporan belum ada atau belum di-approve
		}
		return false, err // Error database
	}
	return true, nil // Laporan sudah di-approve!
}

func (r *CampaignRepository) IsDisbursementApproved(campaignID string, phase int) (bool, error) {
	var exists int
	query := `SELECT 1 FROM campaign_disbursements WHERE campaign_id = ? AND phase = ? AND status = 'APPROVED' LIMIT 1`
	err := r.DB.QueryRow(query, campaignID, phase).Scan(&exists)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil // Uang belum cair atau belum di-approve
		}
		return false, err // Error database
	}
	return true, nil // Uang sudah cair!
}

func (r *CampaignRepository) CreateDisbursementRequest(campaignID string, phase int) error {
	query := `
        INSERT INTO campaign_disbursements (campaign_id, phase, status)
        VALUES (?, ?, 'PENDING')
    `
	// Perhatikan: Tidak ada deskripsi dan foto di sini
	_, err := r.DB.Exec(query, campaignID, phase)
	return err
}

func (r *CampaignRepository) CreateReport(input model.CampaignReportInput) error {
	query := `
        INSERT INTO campaign_reports (campaign_id, phase, description, proof_images, status)
        VALUES (?, ?, ?, ?, 'PENDING')
    `
	_, err := r.DB.Exec(query, input.CampaignID, input.Phase, input.Description, input.ProofURL)
	return err
}

// internal/repository/campaign_repository.go

func (r *CampaignRepository) GetPendingDisbursements() ([]model.PendingDisbursementResponse, error) {
	var disbursements []model.PendingDisbursementResponse

	// JOIN antara tabel pencairan dan tabel detail kampanye
	query := `
		SELECT 
			cd.id, 
			cd.campaign_id, 
			c.title AS campaign_title, 
			c.wallet_address, 
			cd.phase, 
			cd.status, 
			cd.created_at
		FROM campaign_disbursements cd
		JOIN campaigns c ON cd.campaign_id = c.id
		WHERE cd.status = 'PENDING'
		ORDER BY cd.created_at ASC
	`

	rows, err := r.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Looping untuk memasukkan data ke dalam slice/array
	for rows.Next() {
		var d model.PendingDisbursementResponse
		err := rows.Scan(
			&d.ID,
			&d.CampaignID,
			&d.CampaignTitle,
			&d.WalletAddress,
			&d.Phase,
			&d.Status,
			&d.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		disbursements = append(disbursements, d)
	}

	return disbursements, nil
}

// internal/repository/campaign_repository.go

// A. Mengambil Daftar Antrean Laporan
func (r *CampaignRepository) GetPendingReports() ([]model.PendingReportResponse, error) {
	var reports []model.PendingReportResponse

	// JOIN tabel campaign_reports dan campaigns
	query := `
		SELECT 
			cr.id, 
			cr.campaign_id, 
			c.title AS campaign_title, 
			cr.phase, 
			cr.description,
			cr.proof_images,
			cr.status, 
			cr.created_at
		FROM campaign_reports cr
		JOIN campaigns c ON cr.campaign_id = c.id
		WHERE cr.status = 'PENDING'
		ORDER BY cr.created_at ASC
	`

	rows, err := r.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var r model.PendingReportResponse
		err := rows.Scan(
			&r.ID, &r.CampaignID, &r.CampaignTitle,
			&r.Phase, &r.Description, &r.ProofURL,
			&r.Status, &r.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		reports = append(reports, r)
	}
	return reports, nil
}

// B. Mengambil 1 Laporan untuk divalidasi statusnya
func (r *CampaignRepository) GetReportByID(reportID string) (model.CampaignReport, error) {
	var rep model.CampaignReport
	query := `SELECT id, campaign_id, phase, status FROM campaign_reports WHERE id = ?`
	err := r.DB.QueryRow(query, reportID).Scan(&rep.ID, &rep.CampaignID, &rep.Phase, &rep.Status)
	return rep, err
}

// C. Mengubah status Laporan (Approve/Reject)
func (r *CampaignRepository) UpdateReportStatus(reportID string, status string) error {
	query := `UPDATE campaign_reports SET status = ? WHERE id = ?`
	_, err := r.DB.Exec(query, status, reportID)
	return err
}

// internal/repository/campaign_repository.go

func (r *CampaignRepository) GetTrackingHistory(campaignID string) ([]model.StepperItem, error) {
	var history []model.StepperItem

	query := `
		SELECT 'PENCAIRAN' as type, phase, status, updated_at as time, '' as proof_images
		FROM campaign_disbursements WHERE campaign_id = ?
		
		UNION ALL
		
		SELECT 'LAPORAN' as type, phase, status, updated_at as time, proof_images
		FROM campaign_reports WHERE campaign_id = ?
		
		ORDER BY time DESC
	`

	rows, err := r.DB.Query(query, campaignID, campaignID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var h model.StepperItem

		// Scan sekarang menangkap 5 variabel (termasuk h.ProofImages)
		err := rows.Scan(&h.Type, &h.Phase, &h.Status, &h.Date, &h.ProofImage)
		if err != nil {
			return nil, err
		}

		history = append(history, h)
	}

	return history, nil
}

// Mengambil 1 laporan spesifik berdasarkan ID Kampanye dan Fasenya
func (r *CampaignRepository) GetReportByPhase(campaignID string, phase int) (model.CampaignReport, error) {
	var rep model.CampaignReport
	query := `SELECT id, status FROM campaign_reports WHERE campaign_id = ? AND phase = ? LIMIT 1`

	// Jika data tidak ditemukan, error dari QueryRow biasanya adalah sql.ErrNoRows
	err := r.DB.QueryRow(query, campaignID, phase).Scan(&rep.ID, &rep.Status)
	return rep, err
}

// internal/repository/campaign_repository.go

func (r *CampaignRepository) RejectReport(reportID string, reason string) error {
	query := `UPDATE campaign_reports SET status = 'REJECTED', reject_reason = ? WHERE id = ?`
	_, err := r.DB.Exec(query, reason, reportID)
	return err
}

// internal/repository/campaign_repository.go

func (r *CampaignRepository) UpdateReport(reportID string, description string, jsonURLs string) error {
	// Status dikembalikan ke PENDING agar Admin tahu ada perbaikan yang perlu direview
	query := `UPDATE campaign_reports SET description = ?, proof_images = ?, status = 'PENDING', reject_reason = NULL WHERE id = ?`
	_, err := r.DB.Exec(query, description, jsonURLs, reportID)
	return err
}