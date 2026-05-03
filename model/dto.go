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
	ID              string `json:"id"`
	IdToken         string `json:"id_token" form:"id_token"`
	GoogleUID       string `json:"google_uid"`
	Name            string `json:"full_name" form:"full_name"`
	Email           string `json:"email"`
	WalletAddress   string `json:"wallet_address" form:"wallet_address"`
	Role            string `json:"role" form:"role"`
	Isverified      int    `json:"is_verified"`
	AvatarUrl       string `json:"avatar_url"`
	PhotoProfile    string `json:"photo_profile"`
	BeneficiaryType string `json:"beneficiary_type"`
}

type UserProfile struct {
	ID             string    `json:"id" db:"id"`
	UserID         string    `json:"user_id" db:"user_id"`
	FullName       string    `json:"full_name" db:"full_name"`
	WalletAddress  string    `json:"wallet_address" db:"wallet_address"`
	NIK            string    `json:"nik" db:"nik"`
	NPWP           string    `json:"npwp" db:"npwp"`
	JenisKelamin   string    `json:"jenis_kelamin" db:"jenis_kelamin"`
	Agama          string    `json:"agama" db:"agama"`
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
	User    User               `json:"user" binding:"required"`
	Profile BeneficiaryProfile `json:"profile" binding:"required"`
}

type BeneficiaryProfile struct {
	ID                 string    `json:"id" db:"id"`
	UserID             string    `json:"user_id" db:"user_id"`
	WalletAddress      string    `json:"wallet_address" db:"wallet_address" form:"wallet_address"`
	BeneficiaryType    string    `json:"beneficiary_type" db:"beneficiary_type" form:"beneficiary_type"`
	FullName           string    `json:"full_name" db:"full_name" form:"full_name"`
	PhoneNumber        *string   `json:"phone_number" db:"phone_number" form:"phone_number"`
	Alamat             *string   `json:"alamat" db:"alamat" form:"alamat"`
	BioDescription     *string   `json:"bio_description" db:"bio_description" form:"bio_description"`
	PhotoProfile       string    `json:"photo_profile" db:"photo_profile"`
	AvatarUrl          *string   `json:"avatar_url" db:"avatar_url"`
	Nik                *string   `json:"nik,omitempty" db:"nik" form:"nik"`
	JenisKelamin       *string   `json:"jenis_kelamin,omitempty" db:"jenis_kelamin" form:"jenis_kelamin"`
	Agama              *string   `json:"agama,omitempty" db:"agama" form:"agama"`
	TempatLahir        *string   `json:"tempat_lahir,omitempty" db:"tempat_lahir" form:"tempat_lahir"`
	TanggalLahir       string    `json:"tanggal_lahir" form:"tanggal_lahir"`
	Pekerjaan          *string   `json:"pekerjaan,omitempty" db:"pekerjaan" form:"pekerjaan"`
	RegistrationNumber *string   `json:"registration_number,omitempty" db:"registration_number" form:"registration_number"`
	Npwp               *string   `json:"npwp,omitempty" db:"npwp" form:"npwp"`
	CreatedAt          time.Time `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time `json:"updated_at" db:"updated_at"`
}

type RefreshToken struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
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
	ID            string    `json:"id"`
	UserID        string    `json:"user_id"`
	WalletAddress string    `json:"wallet_address" form:"wallet_address"`
	FullName      string    `json:"full_name" form:"full_name"`
	CategoryID    int       `json:"category_id" form:"category_id"`
	Title         string    `json:"title" form:"title"`
	Slug          string    `json:"slug"`
	Description   string    `json:"description" form:"description"`
	Story         string    `json:"story" form:"story"` // Field baru
	TargetAmount  float64   `json:"target_amount" form:"target_amount"`
	CurrentAmount float64   `json:"current_amount"`
	ImageBanner   string    `json:"image_banner"`
	EndDate       *string   `json:"end_date" form:"end_date"`
	Status        string    `json:"status"`
	ApprovedBy    string    `json:"approved_by"`
	ApprovedAt    time.Time `json:"approved_at"`
	CreatedAt     time.Time `json:"created_at"`
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

type AlchemyWebhookPayload struct {
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

type APIResponse struct {
	Error   bool        `json:"error"`          // True/False
	Message string      `json:"message"`        // Pesan untuk user
	Type    string      `json:"type,omitempty"` // Jenis error (ValidationError, etc)
	Data    interface{} `json:"data,omitempty"` // Data sukses (opsional)
}
