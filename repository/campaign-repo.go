package repository

import (
	"context"

	"github.com/muhammadfarrasfajri/filantropi/model"
)

type CampaighRepo interface {
	CreateCampaign(campaign model.Campaign) (model.Campaign, error)
	GetCampaignByStatus(status string) ([]model.Campaign, error)
	GetBySlug(slug string) (model.Campaign, error)
	GetByUserID(userID string) ([]model.Campaign, error)
	GetByID(id string) (model.Campaign, error)
	UpdateCampaign(campaign model.Campaign) (model.Campaign, error)
	ApproveCampaign(campaignID string, adminID string) error
	GetCampaignByWallet(ctx context.Context, walletAddr string) (model.Campaign, error)
}
