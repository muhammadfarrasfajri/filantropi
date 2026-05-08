package service

import (
	"errors"
	"fmt"

	"github.com/muhammadfarrasfajri/filantropi/model"
	"github.com/muhammadfarrasfajri/filantropi/repository"
	"github.com/muhammadfarrasfajri/filantropi/utils"
)

type AdminService struct {
	AdminRepo    repository.AdminRepo
	CampaignRepo repository.CampaignRepo
	UserRepo     repository.UserRepo
}

func NewAdminService(adminRepo repository.AdminRepo, campaignRepo repository.CampaignRepo, userRepo repository.UserRepo) *AdminService {
	return &AdminService{
		AdminRepo:    adminRepo,
		CampaignRepo: campaignRepo,
		UserRepo:     userRepo,
	}
}

func (s *AdminService) GetAllCampaign(status string, search string) ([]model.Campaign, error) {
	campaigns, err := s.AdminRepo.GetAllCampaigns(status, search)
	if err != nil {
		return nil, err
	}
	return campaigns, nil
}

func (s *AdminService) GetAllUsersForAdmin(search string) ([]model.AdminUserList, error) {
	return s.AdminRepo.GetAllUsersForAdmin(search)
}

func (s *AdminService) GetUserDetailForAdmin(userID string) (*model.AdminUserListDetail, error) {
	return s.AdminRepo.GetUserDetailForAdmin(userID)
}

func (s *AdminService) GetDashboardSummary() (model.AdminDashboardSummary, error) {
	return s.AdminRepo.GetDashboardSummary()
}

func (s *AdminService) ApproveCampaign(campaignSlug string, adminID string) error {
	// 1. Cek apakah campaign-nya ada
	campaign, err := s.CampaignRepo.GetBySlug(campaignSlug)
	if err != nil {
		return errors.New("kampanye tidak ditemukan")
	}

	if campaign.Status != "pending" && campaign.Status != "rejected" {
		return errors.New("hanya kampanye dengan status pending yang bisa disetujui")
	}

	// 3. Update status ke active
	campaign.Status = "active"
	campaign.ReviewedBy = adminID

	return s.AdminRepo.ApproveCampaign(campaignSlug, adminID)
}

func (s *AdminService) RejectedCampaign(campaignSlug string, adminID string, rejectedReason string) error {
	// 1. Cek apakah campaign-nya ada
	campaign, err := s.CampaignRepo.GetBySlug(campaignSlug)
	if err != nil {
		return errors.New("kampanye tidak ditemukan")
	}

	if campaign.Status != "pending" {
		return errors.New("hanya kampanye dengan status pending yang bisa direject")
	}

	// 3. Update status ke active
	campaign.Status = "rejected"
	campaign.ReviewedBy = adminID

	return s.AdminRepo.RejectCampaign(campaignSlug, adminID, rejectedReason)
}

func (c *AdminService) UpdateUserVerification(userID string, input model.InputEmaiVerified) (string, error) {

	err := c.AdminRepo.UpdateUserVerification(userID, *input.IsVerified)
	if err != nil {
		return "", err
	}

	if *input.IsVerified == 0 && input.Reason != "" {
		// Ambil email user
		userEmail, errEmail := c.UserRepo.GetUserEmailByID(userID)
		if errEmail == nil {
			// Kirim email penolakan di dalam goroutine agar tidak bikin API lambat/lemot
			go func(email, alasan string) {
				errSend := utils.SendRejectionEmail(email, alasan)
				if errSend != nil {
					fmt.Printf("[ERROR] Gagal kirim email ke %s: %v\n", email, errSend)
				} else {
					fmt.Printf("[SUCCESS] Email penolakan terkirim ke %s\n", email)
				}
			}(userEmail, input.Reason)
		}
	}

	pesanStatus := "Berhasil diverifikasi"
	if *input.IsVerified == 0 {
		pesanStatus = "Verifikasi ditolak. Email pemberitahuan sedang dikirim ke user (jika alasan disertakan)."
	}
	return pesanStatus, nil
}

// internal/service/campaign_service.go

func (s *AdminService) ApproveDisbursement(disbursementID string) error {
	// 1. Cek apakah request pencairan ini benar-benar ada di database
	disbursement, err := s.AdminRepo.GetDisbursementByID(disbursementID)
	if err != nil {
		return errors.New("data pengajuan pencairan tidak ditemukan")
	}

	// 2. Validasi Status: Jangan sampai Admin nge-klik Approve 2 kali
	if disbursement.Status == "APPROVED" {
		return errors.New("pengajuan pencairan ini sudah disetujui sebelumnya")
	}
	if disbursement.Status == "REJECTED" {
		return errors.New("tidak bisa menyetujui pengajuan yang sudah ditolak")
	}

	// 3. (Opsional) Jika di masa depan kamu mau menambahkan logika notifikasi email ke user
	// bahwa "Dana Fase X Anda telah cair", kamu bisa memanggil fungsi util.SendEmail() di sini.

	// 4. Ubah statusnya menjadi APPROVED di database
	return s.AdminRepo.UpdateDisbursementStatus(disbursementID, "APPROVED")
}
