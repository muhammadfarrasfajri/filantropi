package repository

import (
	"database/sql"

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

func (r *RegisterRepository) CreateUser(user model.User) error {
	userID := uuid.New().String()

	tx, err := r.DB.Begin()
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	queryUser := "INSERT INTO users (id, email, wallet_address, role) VALUES (?, ?, ?, ?)"
	_, err = tx.Exec(queryUser, userID, user.Email, user.WalletAddress, user.Role)
	if err != nil {
		return err
	}

	queryProfile := "INSERT INTO user_profiles (user_id, full_name) VALUES (?, ?)"
	_, err = tx.Exec(queryProfile, userID, user.Name)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
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
