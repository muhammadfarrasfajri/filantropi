package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/muhammadfarrasfajri/filantropi/model"
)

type UserRepository struct {
	DB *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{
		DB: db,
	}
}

func (r *UserRepository) FindUserById(id string) (*model.BaseUser, error) {
	query := `SELECT id, email, google_uid, wallet_address, role, is_verified FROM users WHERE id = ?`

	user := model.BaseUser{}
	err := r.DB.QueryRow(query, id).Scan(
		&user.ID,
		&user.Email,
		&user.GoogleUID,
		&user.WalletAddress,
		&user.Role,
		&user.Isverified,
	)
	if err == sql.ErrNoRows {
		return nil, errors.New("user not found")
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) FindDonorsById(userId string) (*model.User, error) {
	// Kita gunakan COALESCE untuk mengambil data dari tabel mana pun yang tersedia
	query := `
    SELECT 
        u.id,               -- 1
        u.google_uid,       -- 2
        u.email,            -- 3
        u.wallet_address,   -- 4
        u.role,             -- 5
        u.is_verified,      -- 6
        COALESCE(p.avatar_url, '') as avatar_url,   -- 7 (Dari tabel users)
        COALESCE(p.full_name, 'User') as full_name, -- 8 (Dari user_profiles)
        COALESCE(p.photo_profile, '') as photo_profile -- 9 (Dari user_profiles)
    FROM users u
    LEFT JOIN user_profiles p ON u.id = p.user_id 
    WHERE u.id = ? AND u.role = 'user' 
    LIMIT 1
`

	user := model.User{}
	err := r.DB.QueryRow(query, userId).Scan(
		&user.ID,            // 1
		&user.GoogleUID,     // 2
		&user.Email,         // 3
		&user.WalletAddress, // 4
		&user.Role,          // 5
		&user.Isverified,    // 6
		&user.AvatarUrl,     // 7
		&user.Name,          // 8
		&user.PhotoProfile,  // 9
	)
	if err == sql.ErrNoRows {
		return nil, err
	}
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *UserRepository) UpdateDonors(ctx context.Context, userID string, walletAddress string, fullName string, photoProfile string) error {
	// 1. Mulai Transaksi
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer tx.Rollback()

	now := time.Now()

	// 2. Update tabel 'users' (kolom wallet_address)
	queryUser := `UPDATE users SET wallet_address = ?, updated_at = ? WHERE id = ?`
	_, err = tx.ExecContext(ctx, queryUser, walletAddress, now, userID)
	if err != nil {
		return fmt.Errorf("gagal update tabel users: %v", err)
	}

	// 3. Update tabel 'user_profiles' (kolom full_name)
	queryProfile := `
	UPDATE user_profiles 
    SET 
        full_name = COALESCE(NULLIF(?, ''), full_name),
        photo_profile = COALESCE(NULLIF(?, ''), photo_profile),
        updated_at = ? 
    WHERE user_id = ?`

	_, err = tx.ExecContext(ctx, queryProfile, fullName, photoProfile, now, userID)
	if err != nil {
		return fmt.Errorf("gagal update tabel user_profiles: %v", err)
	}

	// 4. Jika semua oke, simpan perubahan secara permanen
	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (r *UserRepository) FindBeneficiaryById(userId string) (*model.User, *model.BeneficiaryProfile, error) {
	// 1. Gunakan LEFT JOIN agar user tetap ketemu walaupun profilnya belum ada
	queryBase := `
    SELECT 
        u.id, u.google_uid, u.email, u.wallet_address, u.role, u.is_verified, 
        COALESCE(bp.avatar_url, '') as avatar_url, 
        COALESCE(bp.beneficiary_type, '') as beneficiary_type 
    FROM users u 
    LEFT JOIN beneficiary_profiles bp ON u.id = bp.user_id 
    WHERE u.id = ? AND u.role = 'beneficiary' 
    LIMIT 1`

	user := &model.User{}
	err := r.DB.QueryRow(queryBase, userId).Scan(
		&user.ID, &user.GoogleUID, &user.Email, &user.WalletAddress,
		&user.Role, &user.Isverified, &user.AvatarUrl, &user.BeneficiaryType,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil, errors.New("user penerima tidak ditemukan")
		}
		return nil, nil, err
	}

	// 2. Siapkan penampung profil
	profile := &model.BeneficiaryProfile{}
	profile.UserID = userId
	profile.BeneficiaryType = user.BeneficiaryType

	// Jika user.BeneficiaryType kosong, berarti profile beneran belum ada.
	// Langsung return user dengan profile kosong agar tidak error di query selanjutnya.
	if user.BeneficiaryType == "" {
		return user, profile, nil
	}

	var queryProfile string
	if user.BeneficiaryType == "individual" {
		queryProfile = `
            SELECT 
                full_name, phone_number, COALESCE(alamat, ''), COALESCE(bio_description, ''), 
                COALESCE(photo_profile, ''), COALESCE(nik, ''), COALESCE(jenis_kelamin, ''), 
                COALESCE(agama, ''), COALESCE(tempat_lahir, ''), tanggal_lahir, COALESCE(pekerjaan, '')
            FROM beneficiary_profiles 
            WHERE user_id = ?`

		err = r.DB.QueryRow(queryProfile, userId).Scan(
			&profile.FullName, &profile.PhoneNumber, &profile.Alamat, &profile.BioDescription,
			&profile.PhotoProfile, &profile.Nik, &profile.JenisKelamin,
			&profile.Agama, &profile.TempatLahir, &profile.TanggalLahir, &profile.Pekerjaan,
		)
	} else {
		queryProfile = `
            SELECT 
                full_name, phone_number, COALESCE(alamat, ''), COALESCE(bio_description, ''), 
                COALESCE(photo_profile, ''), COALESCE(registration_number, ''), COALESCE(npwp, '')
            FROM beneficiary_profiles 
            WHERE user_id = ?`

		err = r.DB.QueryRow(queryProfile, userId).Scan(
			&profile.FullName, &profile.PhoneNumber, &profile.Alamat, &profile.BioDescription,
			&profile.PhotoProfile, &profile.RegistrationNumber, &profile.Npwp,
		)
	}

	// Jika di sini ErrNoRows, abaikan saja karena artinya profil memang belum lengkap isinya
	if err != nil && err != sql.ErrNoRows {
		return nil, nil, err
	}

	return user, profile, nil
}

func (r *UserRepository) UpdateProfileBeneficiary(ctx context.Context, userId string, profile model.BeneficiaryProfile) error {
	var query string
	var err error

	// Simpan waktu sekarang agar konsisten di semua tabel
	now := time.Now()

	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 1. UPDATE Tabel Users (Update wallet_address utama)
	queryUser := `UPDATE users SET wallet_address = ?, updated_at = ? WHERE id = ?`
	_, err = tx.ExecContext(ctx, queryUser, profile.WalletAddress, now, userId)
	if err != nil {
		return err
	}

	// 3. UPDATE Tabel beneficiary_profiles (Individual vs Organization)
	if profile.BeneficiaryType == "individual" {
		query = `
        UPDATE beneficiary_profiles 
        SET full_name = ?, phone_number = ?, alamat = ?, bio_description = ?, 
            photo_profile = ?, nik = ?, jenis_kelamin = ?, agama = ?, 
            tempat_lahir = ?, tanggal_lahir = ?, pekerjaan = ?, updated_at = ?
        WHERE user_id = ?`

		_, err = tx.ExecContext(ctx, query,
			profile.FullName, profile.PhoneNumber, profile.Alamat, profile.BioDescription,
			profile.PhotoProfile, profile.Nik, profile.JenisKelamin, profile.Agama,
			profile.TempatLahir, profile.TanggalLahir, profile.Pekerjaan, now,
			userId,
		)
	} else {
		query = `
        UPDATE beneficiary_profiles 
        SET full_name = ?, phone_number = ?, alamat = ?, bio_description = ?, 
            photo_profile = ?, registration_number = ?, npwp = ?, updated_at = ?
        WHERE user_id = ?`

		_, err = tx.ExecContext(ctx, query,
			profile.FullName, profile.PhoneNumber, profile.Alamat, profile.BioDescription,
			profile.PhotoProfile, profile.RegistrationNumber, profile.Npwp, now,
			userId,
		)
	}

	if err != nil {
		return err
	}

	return tx.Commit()
}
