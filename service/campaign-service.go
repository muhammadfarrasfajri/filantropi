package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/muhammadfarrasfajri/filantropi/model"
	"github.com/muhammadfarrasfajri/filantropi/repository"
	"github.com/muhammadfarrasfajri/filantropi/utils"
)

type CampaignService struct {
	CampaignRepo repository.CampaignRepo
	UserRepo     repository.UserRepo
}

func NewCampaignService(campaignRepo repository.CampaignRepo, userRepo repository.UserRepo) *CampaignService {
	return &CampaignService{
		CampaignRepo: campaignRepo,
		UserRepo:     userRepo,
	}
}

func (s *CampaignService) CreateCampaign(id string, campaign model.Campaign) (*model.Campaign, error) {
	// 1. LOG AWAL PROSES
	log.Printf("[CreateCampaign] Memulai pembuatan campaign: Title=%s, UserID=%s", campaign.Title, id)

	// Pastikan UserID sinkron dengan parameter id
	campaign.UserID = id

	// 2. AMBIL DATA USER (JANGAN ABAIKAN ERRORNYA)
	beneficiary, profile, err := s.UserRepo.FindBeneficiaryById(id)
	if err != nil {
		log.Printf("[CreateCampaign][Err] User pembuat kampanye tidak ditemukan: %s", id)
		return nil, errors.New("user tidak valid atau tidak ditemukan")
	}

	// 3. LOGIKA PENENTUAN WALLET
	// Jika user adalah individual, PAKSA gunakan wallet dari profilnya
	if profile.BeneficiaryType == "individual" {
		campaign.WalletAddress = beneficiary.WalletAddress
	}

	// 4. VALIDASI WALLET (Dilakukan SETELAH penentuan wallet di atas)
	if campaign.WalletAddress != "" {
		_, err := s.CampaignRepo.GetCampaignByWallet(context.Background(), campaign.WalletAddress)
		if err == nil {
			log.Printf("[CreateCampaign][ValidationErr] Wallet address sudah terdaftar: %s", campaign.WalletAddress)
			return nil, errors.New("wallet address sudah terdaftar untuk kampanye lain")
		}
	}

	// 5. GENERATE DATA OTOMATIS
	campaign.ID = uuid.New().String()
	campaign.Slug = utils.GenerateSlug(campaign.Title)
	campaign.Status = "pending"
	campaign.CurrentAmount = 0

	// 6. LOGIKA BISNIS (VALIDASI TARGET DONASI)
	if campaign.TargetAmount <= 0 {
		log.Printf("[CreateCampaign][ValidationErr] Target donasi tidak valid: %d (ID: %s)", campaign.TargetAmount, campaign.ID)
		return nil, errors.New("target donasi harus lebih dari 0")
	}

	// 7. SIMPAN KE DATABASE
	log.Printf("[CreateCampaign][DB] Mencoba menyimpan campaign ke database. ID: %s", campaign.ID)
	data, err := s.CampaignRepo.CreateCampaign(campaign)
	if err != nil {
		log.Printf("[CreateCampaign][DBErr] Gagal menyimpan ke database: %v. Data: %+v", err, campaign)
		return nil, err
	}

	// 8. SINKRONISASI ALCHEMY DI BACKGROUND (GOROUTINE)
	if campaign.WalletAddress != "" {
		go func() {
			errWebhook := utils.AddWalletToAlchemyWebhook(campaign.WalletAddress)
			if errWebhook != nil {
				// Cukup print log saja, jangan hentikan proses karena data sudah sukses masuk DB
				log.Printf("[ERROR] Gagal auto-register webhook Alchemy untuk kampanye %s: %v\n", campaign.ID, errWebhook)
			}
		}()
	}

	// 9. LOG SUCCESS
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
	existingCampaign.ReviewedAt = time.Time{}
	existingCampaign.ReviewedBy = ""
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

func (s *CampaignService) UpdateWalletAddress(campaignID string, newWalletAddress string) error {

	// 1. AMBIL DATA LAMA DULU DARI DATABASE SEBELUM DI-UPDATE
	campaign, err := s.CampaignRepo.GetByID(campaignID)
	if err != nil {
		return err
	}
	// Simpan dompet lama ke variabel agar tidak hilang
	oldWalletAddress := campaign.WalletAddress

	// 2. SEKARANG BARU UPDATE DATABASE LOKAL DENGAN DOMPET BARU
	err = s.CampaignRepo.UpdateWalletAddress(campaignID, newWalletAddress)
	if err != nil {
		// Jika update DB gagal, batalkan semuanya
		return err
	}

	// 3. JALANKAN LOGIKA ALCHEMY DI BACKGROUND (GOROUTINE)
	// Gunakan 1 Goroutine saja yang membungkus logika pengecekan IF-ELSE
	go func() {
		var errWebhook error

		if oldWalletAddress == "" {
			// Kasus A: Kampanye sebelumnya belum punya dompet (Add)
			errWebhook = utils.AddWalletToAlchemyWebhook(newWalletAddress)
		} else {
			// Kasus B: Kampanye sudah punya dompet (Swap)
			errWebhook = utils.SwapWalletInAlchemyWebhook(oldWalletAddress, newWalletAddress)
		}

		// Tangkap error jika ada kegagalan komunikasi dengan Alchemy
		if errWebhook != nil {
			fmt.Printf("[ERROR] Gagal sinkronisasi dompet ke Alchemy untuk kampanye %s: %v\n", campaignID, errWebhook)
		}
	}()

	// 4. Langsung kembalikan respons SUKSES ke Frontend!
	return nil
}

func (s *CampaignService) CreateDisbursementRequest(campaignID string) error {
	// 1. Cek apakah ada request pencairan yang masih PENDING
	hasPending, err := s.CampaignRepo.HasPendingDisbursement(campaignID)
	if err != nil {
		return err
	}
	if hasPending {
		return errors.New("masih ada pengajuan pencairan yang sedang diproses admin")
	}

	// 2. Hitung ini fase ke-berapa (Berdasarkan jumlah yang sudah di-Approve)
	approvedCount, err := s.CampaignRepo.CountApprovedDisbursements(campaignID)
	if err != nil {
		return err
	}

	requestPhase := approvedCount + 1

	if requestPhase > 5 {
		return errors.New("seluruh dana kampanye (5 fase) sudah dicairkan")
	}

	// 3. ATURAN PENTING: Kalau mau minta uang Fase 2, Laporan Fase 1 harus sudah beres!
	if requestPhase > 1 {
		// Kita cek tabel reports, apakah laporan fase sebelumnya sudah APPROVED?
		isReportDone, err := s.CampaignRepo.IsReportApproved(campaignID, requestPhase-1)
		if err != nil || !isReportDone {
			return errors.New("anda belum menyelesaikan laporan nota untuk fase sebelumnya")
		}
	}

	// 4. Jika semua syarat di atas lolos, baru eksekusi ke Database!
	return s.CampaignRepo.CreateDisbursementRequest(campaignID, requestPhase)
}
func (s *CampaignService) CreateReport(input model.CampaignReportInput) error {
	// 1. Cek apakah ada laporan yang masih PENDING (biar user gak spam klik upload)
	hasPending, err := s.CampaignRepo.HasPendingReport(input.CampaignID)
	if err != nil {
		return err
	}
	if hasPending {
		return errors.New("laporan sebelumnya masih menunggu review admin")
	}

	// 2. Tentukan ini Laporan untuk Fase ke-berapa
	approvedReports, err := s.CampaignRepo.CountApprovedReports(input.CampaignID)
	if err != nil {
		return err
	}

	currentReportPhase := approvedReports + 1

	if currentReportPhase > 4 {
		return errors.New("semua laporan kampanye sudah selesai")
	}

	// 3. ATURAN PENTING: User TIDAK BOLEH upload laporan kalau pencairan fase ini belum di-Approve admin!
	isDisbursementApproved, err := s.CampaignRepo.IsDisbursementApproved(input.CampaignID, currentReportPhase)
	if err != nil || !isDisbursementApproved {
		return errors.New("tidak bisa upload laporan karena dana untuk fase ini belum dicairkan oleh admin")
	}

	// Masukkan fase yang benar ke dalam input
	input.Phase = currentReportPhase

	// ==========================================
	// 4. LOGIKA BARU: CEK APAKAH HARUS UPDATE ATAU INSERT BARU
	// ==========================================
	existingReport, err := s.CampaignRepo.GetReportByPhase(input.CampaignID, currentReportPhase)

	// Jika tidak ada error (artinya datanya ketemu di database)
	if err == nil && existingReport.ID != "" {
		// Jika datanya ketemu dan statusnya REJECTED, lakukan UPDATE data yang lama
		if existingReport.Status == "REJECTED" {
			// Memanggil fungsi UpdateReport yang sudah ada di Repository-mu
			return s.CampaignRepo.UpdateReport(existingReport.ID, input.Description, input.ProofURL)
		}

		// Jika statusnya bukan REJECTED (misal ada kebocoran logika), tolak untuk aman
		return errors.New("data laporan untuk fase ini sudah ada dan tidak dapat ditimpa")
	}

	// 5. Jika lolos semua validasi di atas (berarti ini murni upload pertama kali), masukkan data baru!
	return s.CampaignRepo.CreateReport(input)
}

func (s *CampaignService) GetPendingDisbursements() ([]model.PendingDisbursementResponse, error) {
	return s.CampaignRepo.GetPendingDisbursements()
}

// Mengambil semua laporan yang pending
func (s *CampaignService) GetPendingReports() ([]model.PendingReportResponse, error) {
	return s.CampaignRepo.GetPendingReports()
}

// Menyetujui (Approve) laporan
func (s *CampaignService) ApproveReport(reportID string) error {
	// 1. Pastikan laporannya ada di database
	report, err := s.CampaignRepo.GetReportByID(reportID)
	if err != nil {
		return errors.New("data laporan tidak ditemukan")
	}

	// 2. Cegah Double-Approve
	if report.Status == "APPROVED" {
		return errors.New("laporan ini sudah disetujui sebelumnya")
	}
	if report.Status == "REJECTED" {
		return errors.New("tidak bisa menyetujui laporan yang sudah ditolak")
	}

	// 3. Update statusnya menjadi APPROVED
	// Setelah langkah ini berhasil, secara otomatis KUNCI FASE BERIKUTNYA UNTUK USER AKAN TERBUKA!
	return s.CampaignRepo.UpdateReportStatus(reportID, "APPROVED")
}

// internal/service/campaign_service.go

func (s *CampaignService) GetCampaignMilestoneStatus(campaignID string) (model.MilestoneStatus, error) {

	// 1. Cek apakah ada pencairan yang PENDING
	isDisbursementPending, _ := s.CampaignRepo.HasPendingDisbursement(campaignID)
	if isDisbursementPending {
		// Kita hitung ini fase ke-berapa yang lagi pending
		approvedCount, _ := s.CampaignRepo.CountApprovedDisbursements(campaignID)
		return model.MilestoneStatus{
			CurrentPhase:   approvedCount + 1,
			ActionRequired: "WAITING_DISBURSEMENT_APPROVAL",
			Message:        "Pengajuan pencairan Anda sedang diproses oleh Admin.",
		}, nil
	}

	// 2. Cek apakah ada laporan yang PENDING
	isReportPending, _ := s.CampaignRepo.HasPendingReport(campaignID)
	if isReportPending {
		approvedReports, _ := s.CampaignRepo.CountApprovedReports(campaignID)
		return model.MilestoneStatus{
			CurrentPhase:   approvedReports + 1,
			ActionRequired: "WAITING_REPORT_APPROVAL",
			Message:        "Laporan Anda sedang ditinjau oleh Admin.",
		}, nil
	}

	// 3. Jika tidak ada yang PENDING, kita bandingkan jumlah yang sudah di-APPROVE
	approvedDisbursements, _ := s.CampaignRepo.CountApprovedDisbursements(campaignID)
	approvedReports, _ := s.CampaignRepo.CountApprovedReports(campaignID)

	// Kasus A: Jika sudah 4 Fase selesai semua
	if approvedDisbursements == 4 && approvedReports == 4 {
		return model.MilestoneStatus{
			CurrentPhase:   4,
			ActionRequired: "COMPLETED",
			Message:        "Alhamdulillah, seluruh dana telah disalurkan dan dilaporkan.",
		}, nil
	}

	// Kasus B: Jumlah pencairan SAMA DENGAN jumlah laporan (Misal: 0==0, atau 1==1)
	// Artinya: User harus minta pencairan untuk fase berikutnya!
	if approvedDisbursements == approvedReports {
		nextPhase := approvedDisbursements + 1
		return model.MilestoneStatus{
			CurrentPhase:   nextPhase,
			ActionRequired: "REQUEST_DISBURSEMENT",
			Message:        fmt.Sprintf("Silakan ajukan pencairan dana untuk Fase %d", nextPhase),
		}, nil
	}

	// Kasus C: Jumlah pencairan LEBIH BESAR dari jumlah laporan (Misal: 1 > 0)
	// Artinya: Uang sudah cair, tapi user belum upload laporan!
	if approvedDisbursements > approvedReports {
		return model.MilestoneStatus{
			CurrentPhase:   approvedDisbursements,
			ActionRequired: "UPLOAD_REPORT",
			Message:        fmt.Sprintf("Dana Fase %d telah cair. Silakan unggah laporan kegiatan Anda.", approvedDisbursements),
		}, nil
	}

	return model.MilestoneStatus{}, errors.New("terjadi kesalahan pada kalkulasi status")
}

// internal/service/campaign_service.go

func (s *CampaignService) GetCampaignStepper(campaignID string) ([]model.StepperItem, error) {
	// 1. Buat 8 Checkpoint Kosong
	var stepper []model.StepperItem
	for i := 1; i <= 4; i++ {
		stepper = append(stepper, model.StepperItem{
			Phase:       i,
			Type:        "PENCAIRAN",
			Title:       fmt.Sprintf("Pencairan Dana Fase %d", i),
			Status:      "NOT_STARTED",
			Date:        "",
			ProofImages: []string{}, // Kosongkan secara default
		})
		stepper = append(stepper, model.StepperItem{
			Phase:       i,
			Type:        "LAPORAN",
			Title:       fmt.Sprintf("Laporan Bukti Fase %d", i),
			Status:      "NOT_STARTED",
			Date:        "",
			ProofImages: []string{}, // Kosongkan secara default
		})
	}

	// 2. Ambil data asli dari DB
	rawHistory, err := s.CampaignRepo.GetTrackingHistory(campaignID)
	if err != nil {
		return nil, err
	}

	// 3. Timpa status dan proses fotonya
	for _, raw := range rawHistory {
		for i, step := range stepper {
			if step.Phase == raw.Phase && step.Type == raw.Type {
				stepper[i].Status = raw.Status
				stepper[i].Date = raw.Date

				// KHUSUS LAPORAN: Proses fotonya!
				// TAMBAHAN: Foto HANYA akan dibongkar dan dikirim jika statusnya APPROVED
				if raw.Type == "LAPORAN" && raw.ProofImage != "" && raw.Status == "APPROVED" {
					var images []string
					errParse := json.Unmarshal([]byte(raw.ProofImage), &images)
					if errParse == nil {
						stepper[i].ProofImages = images
					} else {
						// Jaga-jaga kalau datanya bukan JSON (hanya 1 URL biasa)
						stepper[i].ProofImages = []string{raw.ProofImage}
					}
				}
			}
		}
	}

	return stepper, nil
}

// internal/service/campaign_service.go

func (s *CampaignService) RejectReport(reportID string, reason string) error {
	// Pastikan laporan ada
	report, err := s.CampaignRepo.GetReportByID(reportID)
	if err != nil {
		return errors.New("laporan tidak ditemukan")
	}

	if report.Status == "APPROVED" {
		return errors.New("tidak bisa menolak laporan yang sudah disetujui")
	}

	return s.CampaignRepo.RejectReport(reportID, reason)
}

func (s *CampaignService) UpdateReport(reportID string, description string, jsonURLs string) error {
	// 1. Pastikan laporan yang mau diedit itu memang ada di database
	report, err := s.CampaignRepo.GetReportByID(reportID)
	if err != nil {
		return errors.New("data laporan tidak ditemukan")
	}

	// 2. (Opsional) Tambahkan validasi keamanan
	// Misalnya: Laporan yang sudah APPROVED tidak boleh diedit lagi agar data tidak dimanipulasi
	if report.Status == "APPROVED" {
		return errors.New("laporan yang sudah disetujui tidak dapat diubah kembali")
	}

	// 3. Panggil Repository untuk melakukan update data
	errUpdate := s.CampaignRepo.UpdateReport(reportID, description, jsonURLs)
	if errUpdate != nil {
		return errors.New("terjadi kesalahan saat memperbarui laporan di database")
	}

	return nil
}
