package service

import (
	"context"
	"fmt"
	"time"

	"firebase.google.com/go/auth"
	"github.com/google/uuid"
	"github.com/muhammadfarrasfajri/filantropi/middleware"
	"github.com/muhammadfarrasfajri/filantropi/model"
	"github.com/muhammadfarrasfajri/filantropi/repository"
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

func (s *RegisterService) RegisterGoogle(user model.User, profile model.BeneficiaryProfile) (map[string]interface{}, error) {
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
	fmt.Println(profile.PhotoProfile)
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
	if user.Role == "beneficiary" {
		// Jika role-nya penerima manfaat
		err = s.RegisRepo.CreateBeneficiary(user, profile, tokenData)
	} else {
		// Jika role-nya donor (pemberi) atau role lainnya
		err = s.RegisRepo.CreateUser(user, tokenData)
	}

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
