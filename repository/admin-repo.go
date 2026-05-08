package repository

import "github.com/muhammadfarrasfajri/filantropi/model"

type AdminRepo interface {
	GetAllCampaigns(status string, search string) ([]model.Campaign, error)
	ApproveCampaign(campaignSlug string, adminID string) error
	GetAllUsersForAdmin(search string) ([]model.AdminUserList, error)
	GetUserDetailForAdmin(userID string) (*model.AdminUserListDetail, error)
	GetDashboardSummary() (model.AdminDashboardSummary, error)
	UpdateUserVerification(userID string, isVerified int) error
	RejectCampaign(campaignSlug string, adminID string, rejectedReason string) error
	GetDisbursementByID(disbursementID string) (model.CampaignDisbursement, error)
	UpdateDisbursementStatus(disbursementID string, status string) error
}
