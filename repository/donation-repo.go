package repository

import (
	"context"

	"github.com/muhammadfarrasfajri/filantropi/model"
)

type DonationRepo interface {
	CreateDonation(donation model.Donation) (model.Donation, error)
	UpdateSuccess(txHash string, campaignID string, amount float64) error
	UpdateToFailed(txHash string) error
	GetHistoryByUserID(ctx context.Context, userID string) ([]model.DonationHistoryResponse, error)
	GetUserWallet(userID string) (string, error)
	GetCampaignByWallet(walletAddr string) (model.Campaign, error)
}
