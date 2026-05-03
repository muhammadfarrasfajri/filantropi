package repository

import (
	"github.com/muhammadfarrasfajri/filantropi/model"
)

type RefreshTokenRepo interface {
	FindRefreshTokenUser(userID string) (*model.RefreshToken, error)
	UpsertTokenLogin(rt model.RefreshToken) error
	DeleteRefreshToken(token string) error
}
