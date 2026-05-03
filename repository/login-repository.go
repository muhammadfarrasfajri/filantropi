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
