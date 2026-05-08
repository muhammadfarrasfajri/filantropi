package repository

import "github.com/muhammadfarrasfajri/filantropi/model"

type LoginRepo interface {
	FindUserByEmail(email string) (*model.User, error)
	GetAdminByEmail(email string) (model.AdminLogin, error)
	UpdateAdminGoogleUID(adminID string, googleUID string) error
}
