package repository

import "github.com/muhammadfarrasfajri/filantropi/model"

type RegisterRepo interface {
	CreateUser(user model.User, refresh model.RefreshToken) error
	CreateBeneficiary(user model.RegisterBeneficiaryReq, refresh model.RefreshToken) error
	IsEmailExists(email string) (bool, error)
	IsWalletAddressExists(walletAddress string) (bool, error)
}
