package repository

import (
	"context"

	"github.com/muhammadfarrasfajri/filantropi/model"
)

type CampaignRepo interface {
	CreateCampaign(campaign model.Campaign) (model.Campaign, error)
	GetCampaignByStatus(status string) ([]model.Campaign, error)
	GetBySlug(slug string) (model.Campaign, error)
	GetByUserID(userID string) ([]model.Campaign, error)
	GetByID(id string) (model.Campaign, error)
	UpdateCampaign(campaign model.Campaign) (model.Campaign, error)
	GetCampaignByWallet(ctx context.Context, walletAddr string) (model.Campaign, error)
	CreateReport(input model.CampaignReportInput) error
	CreateDisbursementRequest(campaignID string, phase int) error
	CountApprovedDisbursements(campaignID string) (int, error)
	CountApprovedReports(campaignID string) (int, error)
	HasPendingDisbursement(campaignID string) (bool, error)
	UpdateWalletAddress(campaignID string, newWalletAddress string) error
	HasPendingReport(campaignID string) (bool, error)
	IsDisbursementApproved(campaignID string, phase int) (bool, error)
	IsReportApproved(campaignID string, phase int) (bool, error)
	GetPendingDisbursements() ([]model.PendingDisbursementResponse, error)
	GetPendingReports() ([]model.PendingReportResponse, error)
	GetReportByID(reportID string) (model.CampaignReport, error)
	UpdateReportStatus(reportID string, status string) error
	GetTrackingHistory(campaignID string) ([]model.StepperItem, error)
	RejectReport(reportID string, reason string) error
	UpdateReport(reportID string, description string, jsonURLs string) error
	GetReportByPhase(campaignID string, phase int) (model.CampaignReport, error)
}
