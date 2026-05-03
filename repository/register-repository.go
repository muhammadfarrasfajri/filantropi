package repository

import (
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/muhammadfarrasfajri/filantropi/model"
)

type RegisterRepository struct {
	DB *sql.DB
}

func NewRegisterRepository(db *sql.DB) *RegisterRepository {
	return &RegisterRepository{
		DB: db,
	}
}

func (r *RegisterRepository) CreateUser(user model.User, refresh model.RefreshToken) error {
	// 1. Generate ID baru
	userID := uuid.New().String()

	// 2. Mulai Transaksi
	tx, err := r.DB.Begin()
	if err != nil {
		return err
	}

	defer tx.Rollback()

	// 3. Insert User
	queryUser := "INSERT INTO users (id, email, google_uid, wallet_address, role) VALUES (?, ?, ?, ?, ?)"
	_, err = tx.Exec(queryUser, userID, user.Email, user.GoogleUID, user.WalletAddress, user.Role)
	if err != nil {
		return err
	}

	// 4. Insert Profile
	queryProfile := "INSERT INTO user_profiles (user_id, full_name, photo_profile, avatar_url) VALUES (?, ?, ?, ?)"
	_, err = tx.Exec(queryProfile, userID, user.Name, user.PhotoProfile, user.AvatarUrl)
	if err != nil {
		return err
	}

	// 5. Insert Refresh Token (Pakai tx.Exec dan pakai variabel userID!)
	queryRefresh := `
        INSERT INTO refresh_tokens (user_id, token, expires_at) 
        VALUES (?, ?, ?) 
        ON DUPLICATE KEY UPDATE token = VALUES(token), expires_at = VALUES(expires_at)`

	// PERBAIKAN DI SINI
	_, err = tx.Exec(queryRefresh, userID, refresh.Token, refresh.ExpiresAt)
	if err != nil {
		return err
	}

	// 6. Simpan Permanen (Commit)
	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}
func (r *RegisterRepository) CreateBeneficiary(user model.User, profiles model.BeneficiaryProfile, refresh model.RefreshToken) error {
	userID := user.ID
	if userID == "" {
		userID = uuid.New().String()
	}

	tx, err := r.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 1. Insert ke tabel users
	queryUser := "INSERT INTO users (id, email, google_uid, wallet_address, role) VALUES (?, ?, ?, ?, ?)"
	_, err = tx.Exec(queryUser, userID, user.Email, user.GoogleUID, user.WalletAddress, user.Role)
	if err != nil {
		return fmt.Errorf("gagal insert users: %v", err)
	}

	// 2. Tentukan Tipe & ID Profil
	beneficiaryType := user.BeneficiaryType
	if beneficiaryType == "" {
		beneficiaryType = "individual"
	}
	beneficiaryID := uuid.New().String()

	// 3. LOGIKA PERCABANGAN (Pilih salah satu query)
	if beneficiaryType == "individual" {
		queryProfileIndividual := `
			INSERT INTO beneficiary_profiles (
				id, user_id, beneficiary_type, full_name, phone_number, 
				alamat, bio_description, photo_profile, avatar_url, 
				nik, jenis_kelamin, agama, tempat_lahir, tanggal_lahir, pekerjaan
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

		_, err = tx.Exec(queryProfileIndividual,
			beneficiaryID, userID, beneficiaryType, user.Name, profiles.PhoneNumber,
			profiles.Alamat, profiles.BioDescription, profiles.PhotoProfile, user.AvatarUrl,
			profiles.Nik, profiles.JenisKelamin, profiles.Agama, profiles.TempatLahir, profiles.TanggalLahir, profiles.Pekerjaan,
		)
	} else {
		queryProfileOrganization := `
			INSERT INTO beneficiary_profiles (
				id, user_id, beneficiary_type, full_name, phone_number, 
				alamat, bio_description, photo_profile, avatar_url, 
				registration_number, npwp
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

		_, err = tx.Exec(queryProfileOrganization,
			beneficiaryID, userID, beneficiaryType, user.Name, profiles.PhoneNumber,
			profiles.Alamat, profiles.BioDescription, profiles.PhotoProfile, user.AvatarUrl,
			profiles.RegistrationNumber, profiles.Npwp,
		)
	}

	if err != nil {
		return fmt.Errorf("gagal insert beneficiary_profiles (%s): %v", beneficiaryType, err)
	}

	// 4. Insert Refresh Token
	queryRefresh := `
		INSERT INTO refresh_tokens (user_id, token, expires_at) 
		VALUES (?, ?, ?) 
		ON DUPLICATE KEY UPDATE token = VALUES(token), expires_at = VALUES(expires_at)`

	_, err = tx.Exec(queryRefresh, userID, refresh.Token, refresh.ExpiresAt)
	if err != nil {
		return fmt.Errorf("gagal insert refresh_token: %v", err)
	}

	return tx.Commit()
}

func (r *RegisterRepository) IsEmailExists(email string) (bool, error) {
	query := `SELECT 1 FROM users WHERE email = ? LIMIT 1`

	var exists int
	err := r.DB.QueryRow(query, email).Scan(&exists)

	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return true, nil
}

func (r *RegisterRepository) IsWalletAddressExists(walletAddress string) (bool, error) {
	query := `SELECT 1 FROM users WHERE wallet_address = ? LIMIT 1`

	var exists int
	err := r.DB.QueryRow(query, walletAddress).Scan(&exists)

	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return true, nil
}
