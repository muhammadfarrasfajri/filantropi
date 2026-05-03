package repository

import (
	"context"

	"github.com/muhammadfarrasfajri/filantropi/model"
)

type UserRepo interface {
	FindUserById(id string) (*model.BaseUser, error)
	FindDonorsById(userId string) (*model.User, error)
	FindBeneficiaryById(userId string) (*model.User, *model.BeneficiaryProfile, error)
	UpdateDonors(ctx context.Context, userID string, walletAddress string, fullName string, photoProfile string) error
	UpdateProfileBeneficiary(ctx context.Context, userId string, profile model.BeneficiaryProfile) error
}
