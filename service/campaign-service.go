package service

import (
	"errors"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/muhammadfarrasfajri/filantropi/model"
	"github.com/muhammadfarrasfajri/filantropi/repository"
	"github.com/muhammadfarrasfajri/filantropi/utils"
)

type CampaignService struct {
	CampaignRepo repository.CampaighRepo
	UserRepo     repository.UserRepo
}

func NewCampaignService(campaignRepo repository.CampaighRepo, userRepo repository.UserRepo) *CampaignService {
	return &CampaignService{
		CampaignRepo: campaignRepo,
		UserRepo:     userRepo,
	}
}

func (s *CampaignService) CreateCampaign(campaign model.Campaign) (*model.Campaign, error) {
	// 1. LOG AWAL PROSES
	log.Printf("[CreateCampaign] Memulai pembuatan campaign: Title=%s, UserID=%s", campaign.Title, campaign.UserID)

	// GENERATE DATA OTOMATIS
	campaign.ID = uuid.New().String()
	campaign.Slug = utils.GenerateSlug(campaign.Title)
	campaign.Status = "pending"
	campaign.CurrentAmount = 0

	// 2. LOGIKA BISNIS (VALIDASI)
	if campaign.TargetAmount <= 0 {
		log.Printf("[CreateCampaign][ValidationErr] Target donasi tidak valid: %d (ID: %s)", campaign.TargetAmount, campaign.ID)
		return nil, errors.New("target donasi harus lebih dari 0")
	}

	// 3. LOG SEBELUM SIMPAN KE DB
	log.Printf("[CreateCampaign][DB] Mencoba menyimpan campaign ke database. ID: %s", campaign.ID)

	// 4. SIMPAN KE DATABASE
	data, err := s.CampaignRepo.CreateCampaign(campaign)
	if err != nil {
		// LOG ERROR DATABASE
		log.Printf("[CreateCampaign][DBErr] Gagal menyimpan ke database: %v. Data: %+v", err, campaign)
		return nil, err
	}

	// 5. LOG SUCCESS
	log.Printf("[CreateCampaign][Success] Campaign berhasil dibuat! ID: %s, Slug: %s", data.ID, data.Slug)

	return &data, nil
}

func (s *CampaignService) GetCampaignByStatus(status string) ([]model.Campaign, error) {
	// 1. LOG AWAL: Mengetahui parameter apa yang masuk
	log.Printf("[GetCampaignByStatus] Memproses pencarian campaign dengan status: %s", status)

	// Validasi Input (Business Logic)
	if status == "" {
		// LOG PERINGATAN: Input tidak valid
		log.Println("[GetCampaignByStatus][ValidationErr] Parameter status kosong")
		return nil, errors.New("status tidak boleh kosong")
	}

	// 2. LOG SEBELUM KE DB: Memastikan flow sampai ke layer repo
	log.Printf("[GetCampaignByStatus][DB] Mengambil data dari repository untuk status: %s", status)
	result, err := s.CampaignRepo.GetCampaignByStatus(status)

	// 3. Handle Error dari DB
	if err != nil {
		// LOG ERROR: Mencatat detail error dari database
		log.Printf("[GetCampaignByStatus][DBErr] Gagal mengambil data untuk status '%s': %v", status, err)
		return nil, err
	}

	// 4. LOG SUCCESS: Mengetahui jumlah data yang berhasil ditarik
	log.Printf("[GetCampaignByStatus][Success] Berhasil menemukan %d campaign dengan status: %s", len(result), status)

	return result, nil
}

func (s *CampaignService) GetCampaignBySlug(slug string) (model.Campaign, error) {
	// 1. LOG AWAL: Melacak akses berdasarkan slug (bagus untuk SEO/Analytics)
	log.Printf("[GetCampaignBySlug] Mencari campaign dengan slug: %s", slug)

	result, err := s.CampaignRepo.GetBySlug(slug)
	if err != nil {
		// 2. LOG ERROR: Penting untuk tahu jika ada link rusak (404)
		log.Printf("[GetCampaignBySlug][NotFound] Campaign tidak ditemukan untuk slug: %s. Error: %v", slug, err)
		return result, err
	}

	// 3. LOG SUCCESS
	log.Printf("[GetCampaignBySlug][Success] Berhasil memuat campaign: %s (ID: %s)", result.Title, result.ID)
	return result, nil
}

func (s *CampaignService) GetUserCampaigns(userID string) ([]model.Campaign, error) {
	// 1. LOG AWAL: Melacak aktivitas user di dashboard mereka
	log.Printf("[GetUserCampaigns] Mengambil daftar campaign milik UserID: %s", userID)

	result, err := s.CampaignRepo.GetByUserID(userID)
	if err != nil {
		// 2. LOG ERROR
		log.Printf("[GetUserCampaigns][DBErr] Gagal mengambil data untuk UserID: %s. Error: %v", userID, err)
		return nil, err
	}

	// 3. LOG SUCCESS: Mengetahui berapa banyak campaign yang dimiliki user
	log.Printf("[GetUserCampaigns][Success] Ditemukan %d campaign untuk UserID: %s", len(result), userID)
	return result, nil
}

func (s *CampaignService) UpdateCampaign(slug string, userID string, input model.Campaign) (*model.Campaign, error) {
	// 1. LOG AWAL: Melacak siapa yang mencoba mengedit campaign apa
	log.Printf("[UpdateCampaign] Request update dimulai. CampaignID: %s, UserID: %s", slug, userID)

	// Cari dulu campaign-nya di DB
	existingCampaign, err := s.CampaignRepo.GetBySlug(slug)
	if err != nil {
		log.Printf("[UpdateCampaign][NotFound] Campaign tidak ditemukan. ID: %s, Error: %v", slug, err)
		return nil, errors.New("kampanye tidak ditemukan")
	}

	// 2. CEK PEMILIK (Security Check)
	if existingCampaign.UserID != userID {
		// LOG PERINGATAN: Potensi unauthorized access
		log.Printf("[UpdateCampaign][SecurityAlert] User %s mencoba mengedit campaign milik %s!", userID, existingCampaign.UserID)
		return nil, errors.New("anda tidak memiliki akses ke kampanye ini")
	}

	// 3. LOG PROSES MAPPING
	log.Printf("[UpdateCampaign][Data] Memetakan perubahan data untuk CampaignID: %s", slug)

	existingCampaign.WalletAddress = input.WalletAddress
	existingCampaign.CategoryID = input.CategoryID
	existingCampaign.Title = input.Title
	existingCampaign.Description = input.Description
	existingCampaign.Story = input.Story
	existingCampaign.Slug = slug
	existingCampaign.ApprovedAt = time.Time{}
	existingCampaign.ApprovedBy = ""
	existingCampaign.Status = "pending"
	existingCampaign.ImageBanner = input.ImageBanner
	existingCampaign.TargetAmount = input.TargetAmount
	existingCampaign.EndDate = input.EndDate

	if input.ImageBanner != "" {
		log.Printf("[UpdateCampaign][Image] Mendeteksi perubahan banner untuk ID: %s", slug)
		existingCampaign.ImageBanner = input.ImageBanner
	}

	// 4. LOG SEBELUM UPDATE DB
	log.Printf("[UpdateCampaign][DB] Mengirim perintah update ke repository untuk ID: %s", slug)

	updatedData, err := s.CampaignRepo.UpdateCampaign(existingCampaign)
	if err != nil {
		log.Printf("[UpdateCampaign][DBErr] Gagal update ke database untuk ID: %s, Error: %v", slug, err)
		return nil, err
	}

	// 5. LOG SUCCESS
	log.Printf("[UpdateCampaign][Success] Campaign ID: %s berhasil diperbarui oleh User: %s", slug, userID)

	return &updatedData, nil
}

func (s *CampaignService) ApproveCampaign(campaignID string, adminID string) error {
	// 1. Cek apakah campaign-nya ada
	campaign, err := s.CampaignRepo.GetByID(campaignID)
	if err != nil {
		return errors.New("kampanye tidak ditemukan")
	}

	// 2. Cek apakah statusnya memang masih pending
	if campaign.Status != "pending" {
		return errors.New("hanya kampanye dengan status pending yang bisa disetujui")
	}

	// 3. Update status ke active
	campaign.Status = "active"
	campaign.ApprovedBy = adminID

	return s.CampaignRepo.ApproveCampaign(campaignID, adminID)
}
