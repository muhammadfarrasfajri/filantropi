package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"firebase.google.com/go/auth"
	"github.com/google/uuid"
	"github.com/muhammadfarrasfajri/filantropi/middleware"
	"github.com/muhammadfarrasfajri/filantropi/model"
	"github.com/muhammadfarrasfajri/filantropi/repository"
	"github.com/muhammadfarrasfajri/filantropi/utils"
)

type RegisterService struct {
	RegisRepo    repository.RegisterRepo
	RefreshRepo  repository.RefreshTokenRepo
	JWTManager   *middleware.JWTManager
	FirebaseAuth *auth.Client
}

func NewRegistrasiService(regisRepo repository.RegisterRepo, refreshRepo repository.RefreshTokenRepo, jwtManager *middleware.JWTManager, firebaseAuth *auth.Client) *RegisterService {
	return &RegisterService{
		RegisRepo:    regisRepo,
		RefreshRepo:  refreshRepo,
		JWTManager:   jwtManager,
		FirebaseAuth: firebaseAuth,
	}
}

func (s *RegisterService) RegisterGoogleUser(user model.User) (map[string]interface{}, error) {
	ctx := context.Background()
	nowStr := time.Now().Format(time.RFC3339)
	fmt.Println(user.PhotoProfile)
	fmt.Printf("[INFO] [%s] Memulai proses RegisterGoogle untuk email: %s\n", nowStr, user.Email)
	// 1. Verifikasi token id dengan Firebase Auth
	token, err := s.FirebaseAuth.VerifyIDToken(ctx, user.IdToken)
	if err != nil {
		fmt.Printf("[ERROR] [%s] Firebase VerifyIDToken gagal: %v\n", nowStr, err)
		return nil, fmt.Errorf("invalid google token")
	}
	// 2. Cek wallet address (Web3 logic)
	if user.WalletAddress != "" {
		isExistsWallet, err := s.RegisRepo.IsWalletAddressExists(user.WalletAddress)
		if err != nil {
			fmt.Printf("[ERROR] [%s] Gagal cek wallet di DB: %v\n", nowStr, err)
			return nil, err
		}
		if isExistsWallet {
			fmt.Printf("[WARN] [%s] Registrasi ditolak: Wallet %s sudah terdaftar\n", nowStr, user.WalletAddress)
			return nil, fmt.Errorf("wallet address already exists")
		}
	}

	if user.Role == "user" {
		user.Isverified = 1
	} else {
		user.Isverified = 0
	}

	// 3. Set default values & extract claims dari Google
	if user.ID == "" {
		user.ID = uuid.New().String()
	}

	if email, ok := token.Claims["email"].(string); ok {
		user.Email = email
	}
	if pic, ok := token.Claims["picture"].(string); ok {
		user.AvatarUrl = pic
	}
	if user.Name == "" {
		if n, ok := token.Claims["name"].(string); ok {
			user.Name = n
		}
	}
	user.GoogleUID = token.UID

	// 4. Check if email already exists
	isExistsEmail, err := s.RegisRepo.IsEmailExists(user.Email)
	if err != nil {
		fmt.Printf("[ERROR] [%s] Gagal cek email di DB: %v\n", nowStr, err)
		return nil, err
	}
	if isExistsEmail {
		fmt.Printf("[WARN] [%s] Registrasi ditolak: Email %s sudah ada\n", nowStr, user.Email)
		return nil, fmt.Errorf("email already exists")
	}

	// 5. Generate JWT Tokens
	accessToken, err := s.JWTManager.GenerateAccessToken(user.ID, user.Email, user.Role)
	if err != nil {
		fmt.Printf("[ERROR] [%s] Gagal generate Access Token: %v\n", nowStr, err)
		return nil, err
	}

	refreshToken, err := s.JWTManager.GenerateRefreshToken(user.ID)
	if err != nil {
		fmt.Printf("[ERROR] [%s] Gagal generate Refresh Token: %v\n", nowStr, err)
		return nil, err
	}

	refreshTokenHash := middleware.HashToken(refreshToken)
	tokenData := model.RefreshToken{
		UserID:    user.ID,
		Token:     refreshTokenHash,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}

	// 6. LOGIKA PERCABANGAN BERDASARKAN ROLE
	// Jika role-nya donor (pemberi) atau role lainnya
	err = s.RegisRepo.CreateUser(user, tokenData)
	if err != nil {
		fmt.Printf("[ERROR] [%s] Gagal simpan data %s ke database: %v\n", nowStr, user.Role, err)
		return nil, fmt.Errorf("failed to create %s: %w", user.Role, err)
	}
	fmt.Printf("[SUCCESS] [%s] %s berhasil didaftarkan. ID: %s\n", nowStr, user.Role, user.ID)

	return map[string]interface{}{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"user": map[string]interface{}{
			"id":    user.ID,
			"name":  user.Name,
			"email": user.Email,
			"role":  user.Role,
		},
	}, nil
}
func (s *RegisterService) RegisterGoogleBeneficiaries(user model.RegisterBeneficiaryReq) (map[string]interface{}, error) {
	ctx := context.Background()
	nowStr := time.Now().Format(time.RFC3339)
	fmt.Println(user.PhotoProfile)
	fmt.Println(user.BeneficiaryType)
	fmt.Println(user)
	fmt.Printf("[INFO] [%s] Memulai proses RegisterGoogle untuk email: %s\n", nowStr, user.Email)
	cleanPhone := strings.TrimSpace(user.PhoneNumber)
	panjang := len(cleanPhone)
	fmt.Println(panjang)
	if panjang < 11 || panjang > 15 {
		return nil, errors.New("invalid phone number")
	}
	// Checking NIK
	if len(user.Nik) != 16 {
		return nil, errors.New("invalid NIK number")
	}
	// Checking NPWP
	if user.Npwp != nil && *user.Npwp != "" {
		isValid, message := utils.ValidateNPWP(*user.Npwp)
		if !isValid || message != "" {
			return nil, errors.New("invalid NPWP: " + message)
		}
	}

	// 1. Verifikasi token id dengan Firebase Auth
	token, err := s.FirebaseAuth.VerifyIDToken(ctx, user.IdToken)
	if err != nil {
		fmt.Printf("[ERROR] [%s] Firebase VerifyIDToken gagal: %v\n", nowStr, err)
		return nil, fmt.Errorf("invalid google token")
	}
	// 2. Cek wallet address (Web3 logic)
	if user.WalletAddress != "" {
		isExistsWallet, err := s.RegisRepo.IsWalletAddressExists(user.WalletAddress)
		if err != nil {
			fmt.Printf("[ERROR] [%s] Gagal cek wallet di DB: %v\n", nowStr, err)
			return nil, err
		}
		if isExistsWallet {
			fmt.Printf("[WARN] [%s] Registrasi ditolak: Wallet %s sudah terdaftar\n", nowStr, user.WalletAddress)
			return nil, fmt.Errorf("wallet address already exists")
		}
	}

	// 3. Set default values & extract claims dari Google
	if user.ID == "" {
		user.ID = uuid.New().String()
	}

	if email, ok := token.Claims["email"].(string); ok {
		user.Email = email
	}
	if pic, ok := token.Claims["picture"].(string); ok {
		user.AvatarUrl = pic
	}
	if user.FullName == "" {
		if n, ok := token.Claims["name"].(string); ok {
			user.FullName = n
		}
	}
	user.GoogleUID = token.UID

	// 4. Check if email already exists
	isExistsEmail, err := s.RegisRepo.IsEmailExists(user.Email)
	if err != nil {
		fmt.Printf("[ERROR] [%s] Gagal cek email di DB: %v\n", nowStr, err)
		return nil, err
	}
	if isExistsEmail {
		fmt.Printf("[WARN] [%s] Registrasi ditolak: Email %s sudah ada\n", nowStr, user.Email)
		return nil, fmt.Errorf("email already exists")
	}

	// 5. Generate JWT Tokens
	accessToken, err := s.JWTManager.GenerateAccessToken(user.ID, user.Email, user.Role)
	if err != nil {
		fmt.Printf("[ERROR] [%s] Gagal generate Access Token: %v\n", nowStr, err)
		return nil, err
	}

	refreshToken, err := s.JWTManager.GenerateRefreshToken(user.ID)
	if err != nil {
		fmt.Printf("[ERROR] [%s] Gagal generate Refresh Token: %v\n", nowStr, err)
		return nil, err
	}

	refreshTokenHash := middleware.HashToken(refreshToken)
	tokenData := model.RefreshToken{
		UserID:    user.ID,
		Token:     refreshTokenHash,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}

	err = s.RegisRepo.CreateBeneficiary(user, tokenData)
	if err != nil {
		fmt.Printf("[ERROR] [%s] Gagal simpan data %s ke database: %v\n", nowStr, user.Role, err)
		// Jika DB gagal, pendaftaran langsung dibatalkan (Alchemy tidak disentuh sama sekali)
		return nil, fmt.Errorf("failed to create %s: %w", user.Role, err)
	}

	fmt.Printf("[SUCCESS] [%s] %s berhasil didaftarkan. ID: %s\n", nowStr, user.Role, user.ID)

	// 7. SINKRONISASI ALCHEMY DI BACKGROUND (GOROUTINE)
	if user.BeneficiaryType == "individual" && user.WalletAddress != "" {
		go func() {
			errWebhook := utils.AddWalletToAlchemyWebhook(user.WalletAddress)
			if errWebhook != nil {
				// ⚠️ CATATAN PENTING:
				// Karena ini berjalan di background, kita tidak bisa me-return error ke Frontend.
				// Kita hanya bisa mencatat errornya di terminal agar kamu (Backend) tahu kalau ada yang gagal.
				fmt.Printf("[ERROR] Gagal register webhook Alchemy di background: %v\n", errWebhook)
			}
		}()
	}

	// 8. KEMBALIKAN RESPONSE LANGSUNG KE FRONTEND (Sangat Cepat)
	return map[string]interface{}{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"user": map[string]interface{}{
			"id":    user.ID,
			"name":  user.FullName,
			"email": user.Email,
			"role":  user.Role,
		},
	}, nil
}
