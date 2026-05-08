package repository

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/muhammadfarrasfajri/filantropi/model"
)

type AdminRepository struct {
	DB *sql.DB
}

func NewAdminRepository(db *sql.DB) *AdminRepository {
	return &AdminRepository{
		DB: db,
	}
}

// func (r *AdminRepository) LoginAdmin()

func (r *AdminRepository) GetAllCampaigns(status string, search string) ([]model.Campaign, error) {
	var campaigns []model.Campaign

	// Query dasar
	query := `
    SELECT 
        c.id, 
        c.user_id, 
        bp.full_name, 
        c.title, 
        c.category_id, 
        c.wallet_address, 
        c.slug, 
        c.target_amount, 
        c.status, 
        c.created_at 
    FROM campaigns c
    LEFT JOIN beneficiary_profiles bp ON c.user_id = bp.user_id
    WHERE 1=1
`
	var args []interface{}

	// Filter berdasarkan status (jika ada)
	if status != "" {
		query += " AND status = ?"
		args = append(args, status)
	}

	// Filter berdasarkan judul (jika ada search)
	if search != "" {
		query += " AND title LIKE ?"
		args = append(args, "%"+search+"%")
	}

	query += " ORDER BY created_at DESC"

	rows, err := r.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var cp model.Campaign
		rows.Scan(&cp.ID, &cp.UserID, &cp.FullName, &cp.Title, &cp.CategoryID, &cp.WalletAddress, &cp.Slug, &cp.TargetAmount, &cp.Status, &cp.CreatedAt)
		campaigns = append(campaigns, cp)
	}

	return campaigns, nil
}

// internal/repository/user_repository.go

func (r *AdminRepository) GetAllUsersForAdmin(search string) ([]model.AdminUserList, error) {
	var users []model.AdminUserList

	// Query dengan LEFT JOIN ke kedua tabel profil
	query := `
		SELECT 
			u.id, 
			u.email, 
			u.role, 
			u.wallet_address, 
			u.is_verified, 
			u.created_at,
			-- Ambil nama dari penerima, kalau kosong ambil dari user biasa
			COALESCE(bp.full_name, up.full_name, 'Belum Isi Profil') AS full_name,
			-- Ambil no HP dari penerima, kalau kosong ambil dari user biasa
			COALESCE(bp.phone_number,'-') AS phone_number,
			COALESCE(bp.beneficiary_type, '-') AS beneficiary_type
		FROM users u
		LEFT JOIN beneficiary_profiles bp ON u.id = bp.user_id
		LEFT JOIN user_profiles up ON u.id = up.user_id
		WHERE u.deleted_at IS NULL
	`
	var args []interface{}

	// Fitur Search: Cari berdasarkan Email atau Nama
	if search != "" {
		query += ` AND (u.email LIKE ? OR bp.full_name LIKE ? OR up.full_name LIKE ?)`
		searchParam := "%" + search + "%"
		args = append(args, searchParam, searchParam, searchParam)
	}

	query += ` ORDER BY u.created_at DESC`

	rows, err := r.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var u model.AdminUserList
		// Gunakan sql.Null type karena data dari JOIN atau DB bisa bernilai NULL
		var wallet, fullName, phone sql.NullString
		var isVerified sql.NullBool

		err := rows.Scan(
			&u.ID,
			&u.Email,
			&u.Role,
			&wallet,
			&isVerified,
			&u.CreatedAt,
			&fullName,
			&phone,
			&u.BeneficiaryType,
		)

		if err == nil {
			u.WalletAddress = wallet.String
			u.FullName = fullName.String
			u.PhoneNumber = phone.String
			u.IsVerified = isVerified.Bool

			users = append(users, u)
		} else {
			fmt.Printf("[ERROR] Gagal scan row data user: %v\n", err)
		}
	}

	return users, nil
}

func (r *AdminRepository) GetUserDetailForAdmin(userID string) (*model.AdminUserListDetail, error) {
	var u model.AdminUserListDetail

	// Query dengan LEFT JOIN ke kedua tabel profil
	query := `
		SELECT 
			u.id, 
			u.email, 
			u.role, 
			u.wallet_address, 
			u.is_verified, 
			u.created_at,
			COALESCE(bp.full_name, up.full_name, 'Belum Isi Profil') AS full_name,
			COALESCE(bp.phone_number, '-') AS phone_number,
			COALESCE(bp.beneficiary_type, '-') AS beneficiary_type,
			COALESCE(bp.alamat, '-') AS alamat,
			COALESCE(bp.bio_description, '-') AS bio_description,
			COALESCE(bp.photo_profile, up.photo_profile, '-') AS photo_profile,
			COALESCE(bp.nik, '-') AS nik,
			COALESCE(bp.pic, '-') AS pic,
			COALESCE(bp.url_ktp, '-') AS url_ktp,
			COALESCE(bp.npwp, '-') AS npwp,
			COALESCE(bp.registration_number, '-') AS registration_number
		FROM users u
		LEFT JOIN beneficiary_profiles bp ON u.id = bp.user_id
		LEFT JOIN user_profiles up ON u.id = up.user_id
		WHERE u.deleted_at IS NULL AND u.id = ?
	`

	// Gunakan sql.Null type karena data dari JOIN atau DB bisa bernilai NULL
	var wallet, fullName, phone sql.NullString
	var isVerified sql.NullBool

	// Gunakan QueryRow karena kita hanya mengharapkan 1 data detail
	err := r.DB.QueryRow(query, userID).Scan(
		&u.ID,
		&u.Email,
		&u.Role,
		&wallet,
		&isVerified,
		&u.CreatedAt,
		&fullName,
		&phone,
		&u.BeneficiaryType,
		&u.Alamat,
		&u.BioDescription,
		&u.PhotoProfile,
		&u.NIK,
		&u.PIC,
		&u.URLKtp,
		&u.NPWP,
		&u.RegistrationNumber,
	)

	// Handling error jika query gagal atau data tidak ditemukan
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user dengan ID %s tidak ditemukan", userID)
		}
		return nil, err
	}

	// Map data dari tipe sql.Null ke struct utama
	u.WalletAddress = wallet.String
	u.FullName = fullName.String
	u.PhoneNumber = phone.String
	u.IsVerified = isVerified.Bool

	return &u, nil
}

func (r *AdminRepository) GetDashboardSummary() (model.AdminDashboardSummary, error) {
	var summary model.AdminDashboardSummary

	query := `
		SELECT 
			-- Bagian Users
			(SELECT COUNT(CASE WHEN role = 'user' THEN 1 END) FROM users WHERE deleted_at IS NULL) AS total_user,
			(SELECT COUNT(CASE WHEN is_verified = 0 THEN 1 END) FROM users WHERE deleted_at IS NULL) AS total_unverified_user,
			(SELECT COUNT(CASE WHEN is_verified = 1 THEN 1 END) FROM users WHERE deleted_at IS NULL) AS total_verified_user,
			(SELECT COUNT(CASE WHEN role = 'beneficiary' THEN 1 END) FROM users WHERE deleted_at IS NULL) AS total_beneficiary,
			(SELECT COUNT(CASE WHEN role IN ('user', 'beneficiary') THEN 1 END) FROM users WHERE deleted_at IS NULL) AS total_all_users,
			
			-- Bagian Campaigns
			(SELECT COUNT(id) FROM campaigns WHERE status = 'ACTIVE') AS total_active_campaigns,
			(SELECT COUNT(id) FROM campaigns WHERE status = 'PENDING') AS total_pending_campaigns,
			
			-- TAMBAHAN: Hitung semua campaign tanpa memandang statusnya
			(SELECT COUNT(id) FROM campaigns) AS total_all_campaigns
	`

	// Pastikan urutan Scan ini SAMA PERSIS dengan urutan SELECT di atas (sekarang ada 8 baris)
	err := r.DB.QueryRow(query).Scan(
		&summary.UserAmount,
		&summary.UnverifiedUserAmount,
		&summary.VerifiedUserAmount,
		&summary.BeneficiaryAmount,
		&summary.AllUserAmount,

		&summary.ActiveCampaigns,
		&summary.PendingCampaigns,
		&summary.AllCampaignAmount, // <-- Tangkap data tambahan di sini
	)

	if err != nil {
		fmt.Println("[DEBUG] Error get dashboard summary:", err)
		return summary, err
	}

	return summary, nil
}

func (r *AdminRepository) ApproveCampaign(campaignSlug string, adminID string) error {
	query := `
        UPDATE campaigns 
        SET status = 'active', 
            reviewed_by = ?, 
            reviewed_at = NOW() 
        WHERE slug = ?`

	result, err := r.DB.Exec(query, adminID, campaignSlug)
	if err != nil {
		return err
	}

	// CEK: Apakah ada baris yang beneran di-update?
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return errors.New("Approve failed: campaign slug not found")
	}

	return nil
}

func (r *AdminRepository) RejectCampaign(campaignSlug string, adminID string, rejectedReason string) error {
	query := `
        UPDATE campaigns 
        SET status = 'rejected',
		    reject_reason = ?,
            reviewed_by = ?,
            reviewed_at = NOW() 
        WHERE slug = ?`

	result, err := r.DB.Exec(query, rejectedReason, adminID, campaignSlug)
	if err != nil {
		return err
	}

	// CEK: Apakah ada baris yang beneran di-update?
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return errors.New("Reject failed: campaign slug not found")
	}

	return nil
}

func (r *AdminRepository) UpdateUserVerification(userID string, isVerified int) error {
	// 1. Cek dulu apakah ID user ini benar-benar ada di database
	var exists int
	checkQuery := `SELECT 1 FROM users WHERE id = ? AND deleted_at IS NULL`
	errCheck := r.DB.QueryRow(checkQuery, userID).Scan(&exists)

	if errCheck != nil {
		if errCheck == sql.ErrNoRows {
			// Kalau masuk ke sini, berarti ID-nya memang benar-benar tidak ada
			return errors.New("user tidak ditemukan atau id salah")
		}
		// Kalau error lain (misal koneksi putus)
		return errCheck
	}

	// 2. Kalau ID-nya ada, jalankan proses Update
	updateQuery := `UPDATE users SET is_verified = ? WHERE id = ?`
	_, err := r.DB.Exec(updateQuery, isVerified, userID)

	// Kita tidak perlu lagi mengecek RowsAffected == 0,
	// karena tujuan utamanya (memastikan data bernilai sesuai request) sudah tercapai
	return err
}

// internal/repository/campaign_repository.go

// 1. Ambil data pencairan berdasarkan ID-nya
func (r *AdminRepository) GetDisbursementByID(disbursementID string) (model.CampaignDisbursement, error) {
	var d model.CampaignDisbursement
	query := `SELECT id, campaign_id, phase, status FROM campaign_disbursements WHERE id = ?`

	err := r.DB.QueryRow(query, disbursementID).Scan(&d.ID, &d.CampaignID, &d.Phase, &d.Status)
	return d, err
}

// 2. Update status pencairan menjadi APPROVED (atau REJECTED)
func (r *AdminRepository) UpdateDisbursementStatus(disbursementID string, status string) error {
	query := `UPDATE campaign_disbursements SET status = ? WHERE id = ?`
	_, err := r.DB.Exec(query, status, disbursementID)
	return err
}
