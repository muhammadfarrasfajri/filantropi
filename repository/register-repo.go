package repository

import "github.com/muhammadfarrasfajri/filantropi/model"

type RegisterRepo interface {
	CreateUser(user model.User) error
	IsEmailExists(email string) (bool, error)
	IsWalletAddressExists(walletAddress string) (bool, error)
}
