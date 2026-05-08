package model

import "time"

type BaseUser struct {
	ID            string `json:"id"`
	GoogleUID     string `json:"google_uid"`
	Email         string `json:"email"`
	WalletAddress string `json:"wallet_address"`
	Role          string `json:"role"`
	Isverified    int    `json:"is_verified"`
}

type User struct {
	ID              string    `json:"id"`
	IdToken         string    `json:"id_token" form:"id_token"`
	GoogleUID       string    `json:"google_uid"`
	Name            string    `json:"full_name" form:"full_name"`
	Email           string    `json:"email"`
	WalletAddress   string    `json:"wallet_address" form:"wallet_address"`
	Role            string    `json:"role" form:"role"`
	Isverified      int       `json:"is_verified"`
	AvatarUrl       string    `json:"avatar_url"`
	PhotoProfile    string    `json:"photo_profile"`
	BeneficiaryType string    `json:"beneficiary_type"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
}

type UserProfile struct {
	ID             string    `json:"id" db:"id"`
	UserID         string    `json:"user_id" db:"user_id"`
	FullName       string    `json:"full_name" db:"full_name"`
	WalletAddress  string    `json:"wallet_address" db:"wallet_address"`
	NIK            string    `json:"nik" db:"nik"`
	NPWP           string    `json:"npwp" db:"npwp"`
	JenisKelamin   string    `json:"jenis_kelamin" db:"jenis_kelamin"`
	TempatLahir    string    `json:"tempat_lahir" db:"tempat_lahir"`
	TanggalLahir   time.Time `json:"tanggal_lahir" db:"tanggal_lahir"`
	PhoneNumber    string    `json:"phone_number" db:"phone_number"`
	Alamat         string    `json:"alamat" db:"alamat"`
	Pekerjaan      string    `json:"pekerjaan" db:"pekerjaan"`
	BioDescription string    `json:"bio_description" db:"bio_description"`
	PhotoProfile   string    `json:"photo_profile" db:"photo_profile"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

// model/dto.go
type RegisterBeneficiaryReq struct {
	ID                 string    `json:"id"`
	IdToken            string    `json:"id_token" form:"id_token"`
	GoogleUID          string    `json:"google_uid"`
	FullName           string    `json:"full_name" form:"full_name"`
	Email              string    `json:"email"`
	WalletAddress      string    `json:"wallet_address" db:"wallet_address" form:"wallet_address"`
	Role               string    `json:"role" form:"role"`
	Isverified         int       `json:"is_verified"`
	AvatarUrl          string    `json:"avatar_url"`
	PhotoProfile       string    `json:"photo_profile" db:"photo_profile"`
	BeneficiaryType    string    `json:"beneficiary_type" form:"beneficiary_type"`
	UserID             string    `json:"user_id" db:"user_id"`
	PhoneNumber        string    `json:"phone_number" db:"phone_number" form:"phone_number"`
	Alamat             *string   `json:"alamat" db:"alamat" form:"alamat"`
	BioDescription     *string   `json:"bio_description" db:"bio_description" form:"bio_description"`
	Nik                string    `json:"nik" db:"nik" form:"nik"`
	UrlKTP             string    `json:"url_ktp" db:"url_ktp"`
	PIC                string    `json:"pic" db:"pic" form:"pic"`
	JenisKelamin       *string   `json:"jenis_kelamin,omitempty" db:"jenis_kelamin" form:"jenis_kelamin"`
	TempatLahir        *string   `json:"tempat_lahir,omitempty" db:"tempat_lahir" form:"tempat_lahir"`
	TanggalLahir       string    `json:"tanggal_lahir" form:"tanggal_lahir"`
	Pekerjaan          *string   `json:"pekerjaan,omitempty" db:"pekerjaan" form:"pekerjaan"`
	RegistrationNumber *string   `json:"registration_number,omitempty" db:"registration_number" form:"registration_number"`
	Npwp               *string   `json:"npwp,omitempty" db:"npwp" form:"npwp"`
	CreatedAt          time.Time `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time `json:"updated_at" db:"updated_at"`
}

type BeneficiaryProfile struct {
	ID                 string    `json:"id" db:"id"`
	UserID             string    `json:"user_id" db:"user_id"`
	WalletAddress      string    `json:"wallet_address" db:"wallet_address" form:"wallet_address"`
	BeneficiaryType    string    `json:"beneficiary_type" db:"beneficiary_type" form:"beneficiary_type"`
	FullName           string    `json:"full_name" db:"full_name" form:"full_name"`
	PhoneNumber        string    `json:"phone_number" db:"phone_number" form:"phone_number"`
	Alamat             *string   `json:"alamat" db:"alamat" form:"alamat"`
	BioDescription     *string   `json:"bio_description" db:"bio_description" form:"bio_description"`
	PhotoProfile       string    `json:"photo_profile" db:"photo_profile"`
	AvatarUrl          *string   `json:"avatar_url" db:"avatar_url"`
	Nik                string    `json:"nik" db:"nik" form:"nik"`
	UrlKTP             string    `json:"url_ktp" db:"url_ktp"`
	PIC                string    `json:"pic" db:"pic" form:"pic"`
	JenisKelamin       *string   `json:"jenis_kelamin,omitempty" db:"jenis_kelamin" form:"jenis_kelamin"`
	TempatLahir        *string   `json:"tempat_lahir,omitempty" db:"tempat_lahir" form:"tempat_lahir"`
	TanggalLahir       string    `json:"tanggal_lahir" form:"tanggal_lahir"`
	Pekerjaan          *string   `json:"pekerjaan,omitempty" db:"pekerjaan" form:"pekerjaan"`
	RegistrationNumber *string   `json:"registration_number,omitempty" db:"registration_number" form:"registration_number"`
	Npwp               *string   `json:"npwp,omitempty" db:"npwp" form:"npwp"`
	CreatedAt          time.Time `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time `json:"updated_at" db:"updated_at"`
}

// internal/model/user_model.go

type AdminUserList struct {
	ID              string `json:"id"`
	Email           string `json:"email"`
	FullName        string `json:"full_name"`
	Role            string `json:"role"`
	WalletAddress   string `json:"wallet_address"`
	PhoneNumber     string `json:"phone_number"`
	BeneficiaryType string `json:"beneficiary_type"`
	IsVerified      bool   `json:"is_verified"`
	CreatedAt       string `json:"created_at"`
}
type AdminUserListDetail struct {
	ID                 string `json:"id"`
	Email              string `json:"email"`
	FullName           string `json:"full_name"`
	Role               string `json:"role"`
	WalletAddress      string `json:"wallet_address"`
	PhoneNumber        string `json:"phone_number"`
	BeneficiaryType    string `json:"beneficiary_type"`
	Alamat             string `json:"alamat"`
	BioDescription     string `json:"bio_description"`
	PhotoProfile       string `json:"photo_profile"`
	NIK                string `json:"nik"`
	PIC                string `json:"pic"`
	URLKtp             string `json:"url_ktp"`
	RegistrationNumber string `json:"registration_number"`
	NPWP               string `json:"npwp"`
	IsVerified         bool   `json:"is_verified"`
	CreatedAt          string `json:"created_at"`
}

type RefreshToken struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

type AdminRefreshToken struct {
	ID        string    `json:"id"`
	AdminID   string    `json:"admin_id"`
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

type RefreshTokenReq struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type UpdateDonorsRequest struct {
	FullName      string `form:"full_name"`
	WalletAddress string `form:"wallet_address"`
	PhotoProfile  string
}

type Campaign struct {
	ID              string    `json:"id"`
	UserID          string    `json:"user_id"`
	WalletAddress   string    `json:"wallet_address" form:"wallet_address"`
	FullName        string    `json:"full_name" form:"full_name"`
	CategoryID      int       `json:"category_id" form:"category_id"`
	BeneficiaryType string    `json:"beneficiary_type"`
	Title           string    `json:"title" form:"title"`
	Slug            string    `json:"slug"`
	Description     string    `json:"description" form:"description"`
	Story           string    `json:"story" form:"story"` // Field baru
	TargetAmount    float64   `json:"target_amount" form:"target_amount"`
	CurrentAmount   float64   `json:"current_amount"`
	ImageBanner     string    `json:"image_banner"`
	EndDate         *string   `json:"end_date" form:"end_date"`
	Status          string    `json:"status"`
	RejectReason    *string   `json:"reject_reason,omitempty"`
	ReviewedBy      string    `json:"reviewed_by"`
	ReviewedAt      time.Time `json:"reviewed_at"`
	CreatedAt       time.Time `json:"created_at"`
}

type CampaignReport struct {
	ID          string `json:"id"`
	CampaignID  string `json:"campaign_id"`
	Phase       int    `json:"phase"`
	Description string `json:"description"`
	ProofURL    string `json:"proof_url"`
	Status      string `json:"status"`
	CreatedAt   string `json:"created_at"`
}

type Donation struct {
	ID              string  `json:"id"`
	CampaignID      string  `json:"campaign_id"`
	UserID          string  `json:"user_id"`
	WalletAddress   string  `json:"wallet_address"`
	Amount          float64 `json:"amount"`
	Message         string  `json:"message"`
	Status          string  `json:"status"`
	TransactionHash string  `json:"transaction_hash"`
	IsAnonymous     bool    `json:"is_anonymous"`
}

type DonationInput struct {
	CampaignID      string `json:"campaign_id" binding:"required,uuid"`
	Message         string `json:"message"`
	TransactionHash string `json:"transaction_hash" binding:"required"` // Wajib ada untuk verifikasi
	IsAnonymous     bool   `json:"is_anonymous"`                        // Pastikan tipe data bool
}

type DonationHistoryResponse struct {
	ID           string    `json:"id"`
	TxHash       string    `json:"transaction_hash"`
	Amount       float64   `json:"amount"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"date"`
	CampaignName string    `json:"campaign_name"`
}

type TransactionData struct {
	TxHash string `json:"tx_hash"`
	Date   string `json:"date"`
	Type   string `json:"type"`
	Amount string `json:"amount"`
	FromTo string `json:"from_to"`
}

type AlchemyWebhookPayload struct {
	Event struct {
		Activity []struct {
			From     string  `json:"fromAddress"`
			To       string  `json:"toAddress"`
			Value    float64 `json:"value"`
			Hash     string  `json:"hash"`
			Metadata struct {
				BlockTimestamp string `json:"blockTimestamp"`
			} `json:"metadata"`
		} `json:"activity"`
	} `json:"event"`
}

type AlchemyTransferResponse struct {
	Result struct {
		Transfers []AlchemyTransfer `json:"transfers"`
	} `json:"result"`
}

type AlchemyTransfer struct {
	Hash     string  `json:"hash"`
	From     string  `json:"from"`
	To       string  `json:"to"`
	Value    float64 `json:"value"` // Hebatnya Alchemy: Value sudah otomatis di-convert dari Wei ke Desimal!
	Metadata struct {
		BlockTimestamp string `json:"blockTimestamp"`
	} `json:"metadata"`
}

type AlchemyWebhookPayloadOld struct {
	Event struct {
		Activity []struct {
			Hash        string  `json:"hash"`
			FromAddress string  `json:"fromAddress"` // Pengirim
			ToAddress   string  `json:"toAddress"`   // Penerima
			Value       float64 `json:"value"`       // Jumlah Token
			Asset       string  `json:"asset"`       // Nama Token
			Category    string  `json:"category"`    // Biasanya "erc20"
		} `json:"activity"`
	} `json:"event"`
}

type TokenBalanceResponse struct {
	Result struct {
		TokenBalances []struct {
			TokenBalance string `json:"tokenBalance"`
		} `json:"tokenBalances"`
	} `json:"result"`
}

// Struct untuk satu baris riwayat donasi
type BeneficiaryHistory struct {
	DonaturAddress string  `json:"donatur_address"`
	Amount         float64 `json:"amount"`
	TxHash         string  `json:"tx_hash"`
	CreatedAt      string  `json:"created_at"`
}

// Struct untuk riwayat dari sisi Donatur
type DonorsHistory struct {
	CampaignWallet string  `json:"campaign_wallet"`
	Amount         float64 `json:"amount"`
	TxHash         string  `json:"tx_hash"`
	CreatedAt      string  `json:"created_at"`
}

// Struct untuk response gabungan (Saldo + Riwayat)
type WalletStatsResponse struct {
	WalletAddress string               `json:"wallet_address"`
	TotalBalance  float64              `json:"total_balance"`
	History       []BeneficiaryHistory `json:"history"`
}

type AdminDashboardSummary struct {
	UserAmount           int `json:"user_amount"`
	UnverifiedUserAmount int `json:"unverified_user_amount"`
	VerifiedUserAmount   int `json:"verified_user_amount"`
	BeneficiaryAmount    int `json:"beneficiary_amount"` // Jumlah penerima manfaat
	AllUserAmount        int `json:"all_user_amount"`    // Total keduanya
	ActiveCampaigns      int `json:"active_campaigns"`   // Kampanye status ACTIVE
	PendingCampaigns     int `json:"pending_campaigns"`
	AllCampaignAmount    int `json:"all_campaign_amount"`
}

type InputEmaiVerified struct {
	IsVerified *int   `json:"is_verified" binding:"required,oneof=0 1"`
	Reason     string `json:"reason"`
}

type AdminLogin struct {
	ID        string `json:"id"`
	IdToken   string `json:"id_token"`
	GoogleUID string `json:"google_uid"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	CreatedAt string `json:"created_at"`
}

type CampaignReportInput struct {
	CampaignID  string `json:"campaign_id" binding:"required"`
	Phase       int    `json:"phase" binding:"required"`
	Description string `json:"description" binding:"required"`
	ProofURL    string `json:"proof_url"`
}

type CampaignDisbursement struct {
	ID         string `json:"id"`
	CampaignID string `json:"campaign_id"`
	Phase      int    `json:"phase"`
	Status     string `json:"status"`
}

// internal/model/disbursement_model.go

type PendingDisbursementResponse struct {
	ID            string `json:"id"`
	CampaignID    string `json:"campaign_id"`
	CampaignTitle string `json:"campaign_title"`
	WalletAddress string `json:"wallet_address"`
	Phase         int    `json:"phase"`
	Status        string `json:"status"`
	CreatedAt     string `json:"created_at"`
}

// internal/model/report_model.go

type PendingReportResponse struct {
	ID            string `json:"id"`
	CampaignID    string `json:"campaign_id"`
	CampaignTitle string `json:"campaign_title"`
	Phase         int    `json:"phase"`
	Description   string `json:"description"`
	ProofURL      string `json:"proof_url"`
	Status        string `json:"status"`
	CreatedAt     string `json:"created_at"`
}

type MilestoneStatus struct {
	CurrentPhase   int    `json:"current_phase"`
	ActionRequired string `json:"action_required"`
	Message        string `json:"message"`
	ProgressAmount int    `json:"progress_amount"`
}

type StepperItem struct {
	Phase       int      `json:"phase"`
	Type        string   `json:"type"` // "PENCAIRAN" atau "LAPORAN"
	Title       string   `json:"title"`
	Status      string   `json:"status"` // "APPROVED", "PENDING", atau "NOT_STARTED"
	Date        string   `json:"date"`   // Kosong jika belum dimulai
	ProofImage  string   `json:"-"`      // String JSON mentah dari DB (disembunyikan)
	ProofImages []string `json:"proof_images"`
}

type APIResponse struct {
	Error   bool        `json:"error"`          // True/False
	Message string      `json:"message"`        // Pesan untuk user
	Type    string      `json:"type,omitempty"` // Jenis error (ValidationError, etc)
	Data    interface{} `json:"data,omitempty"` // Data sukses (opsional)
}
