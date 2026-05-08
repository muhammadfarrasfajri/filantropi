package repository

import (
	"database/sql"
	"fmt"

	"github.com/muhammadfarrasfajri/filantropi/model"
)

type LoginRepository struct {
	DB *sql.DB
}

func NewLoginRepository(db *sql.DB) *LoginRepository {
	return &LoginRepository{
		DB: db,
	}
}

func (r *LoginRepository) FindUserByEmail(email string) (*model.User, error) {
	// 1. Ambil data dasar
	query := "SELECT id, email, wallet_address, role, is_verified FROM users WHERE email = ?"
	user := &model.User{}
	err := r.DB.QueryRow(query, email).Scan(&user.ID, &user.Email, &user.WalletAddress, &user.Role, &user.Isverified)
	if err != nil {
		return nil, err
	}

	// 2. Tentukan tabel secara dinamis
	tableName := "user_profiles"
	if user.Role == "beneficiary" {
		tableName = "beneficiary_profiles"
	}

	// 3. Eksekusi query profil (Cukup tulis sekali saja)
	queryProfile := fmt.Sprintf("SELECT full_name FROM %s WHERE user_id = ?", tableName)
	err = r.DB.QueryRow(queryProfile, user.ID).Scan(&user.Name)
	if err != nil {
		// Logika tambahan: Jika profil belum ada, jangan hentikan login.
		// Berikan nama default agar user tetap bisa masuk.
		user.Name = "User WANAMA"
		return user, nil
	}

	return user, nil
}

// internal/repository/auth_repository.go

func (r *LoginRepository) GetAdminByEmail(email string) (model.AdminLogin, error) {
	var admin model.AdminLogin

	// Cari admin berdasarkan email yang didapat dari Google
	query := `SELECT id, name, email, role, IFNULL(google_uid, '') FROM admins WHERE email = ?`
	err := r.DB.QueryRow(query, email).Scan(&admin.ID, &admin.Name, &admin.Email, &admin.Role, &admin.GoogleUID)

	return admin, err
}

// Fungsi untuk menyimpan google_uid saat admin pertama kali login
func (r LoginRepository) UpdateAdminGoogleUID(adminID string, googleUID string) error {
	query := `UPDATE admins SET google_uid = ? WHERE id = ?`
	_, err := r.DB.Exec(query, googleUID, adminID)
	return err
}
