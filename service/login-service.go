package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"firebase.google.com/go/auth"
	"github.com/muhammadfarrasfajri/filantropi/middleware"
	"github.com/muhammadfarrasfajri/filantropi/model"
	"github.com/muhammadfarrasfajri/filantropi/repository"
)

type LoginService struct {
	UserRepo     repository.LoginRepo
	RefreshRepo  repository.RefreshTokenRepo
	JWTManager   *middleware.JWTManager
	FirebaseAuth *auth.Client
}

func NewLoginService(userRepo repository.LoginRepo, refreshRepo repository.RefreshTokenRepo, jwtManager *middleware.JWTManager, firebaseAuth *auth.Client) *LoginService {
	return &LoginService{
		UserRepo:     userRepo,
		RefreshRepo:  refreshRepo,
		JWTManager:   jwtManager,
		FirebaseAuth: firebaseAuth,
	}
}

func (s *LoginService) LoginGoogle(idToken string) (map[string]interface{}, error) {
	ctx := context.Background()
	now := time.Now().Format("2006-01-02 15:04:05")

	// 1. LOG: Awal Proses (Jangan log idToken-nya karena kepanjangan & sensitif)
	fmt.Printf("[AUTH-LOGIN] [%s] Mencoba login dengan Google ID Token...\n", now)

	// verify idToken with Firebase Auth
	token, err := s.FirebaseAuth.VerifyIDToken(ctx, idToken)
	if err != nil {
		fmt.Printf("[ERROR] [%s] Firebase VerifyIDToken gagal: %v\n", now, err)
		return nil, errors.New("invalid ID token: " + err.Error())
	}

	// Extract email from token claims
	email, ok := token.Claims["email"].(string)
	if !ok || email == "" {
		fmt.Printf("[WARN] [%s] Login ditolak: Email tidak ditemukan dalam token claims\n", now)
		return nil, errors.New("email not found in token claims")
	}
	fmt.Printf("[INFO] [%s] Token Firebase valid. User Email: %s\n", now, email)

	// find user by email
	user, err := s.UserRepo.FindUserByEmail(email)
	if err != nil {
		fmt.Printf("[WARN] [%s] User %s tidak ditemukan di database. Harus registrasi dulu.\n", now, email)
		return nil, errors.New("user not found, please register first")
	}

	// generate access token
	accessToken, err := s.JWTManager.GenerateAccessToken(user.ID, user.Email, user.Role)
	if err != nil {
		fmt.Printf("[ERROR] [%s] Gagal generate Access Token untuk %s: %v\n", now, email, err)
		return nil, err
	}

	// generate refresh token
	refreshToken, err := s.JWTManager.GenerateRefreshToken(user.ID)
	if err != nil {
		fmt.Printf("[ERROR] [%s] Gagal generate Refresh Token untuk %s: %v\n", now, email, err)
		return nil, err
	}

	// Hash refresh token sebelum disimpan di database
	refreshTokenHash := middleware.HashToken(refreshToken)
	expiresAt := time.Now().Add(7 * 24 * time.Hour)

	tokenData := model.RefreshToken{
		UserID:    user.ID,
		Token:     refreshTokenHash,
		ExpiresAt: expiresAt,
	}

	// Upsert token login (Update jika sudah ada, Insert jika belum)
	err = s.RefreshRepo.UpsertTokenLogin(tokenData)
	if err != nil {
		fmt.Printf("[ERROR] [%s] Gagal simpan/update Refresh Token di DB: %v\n", now, err)
		return nil, err
	}

	// 8. LOG: Berhasil Login
	fmt.Printf("[SUCCESS] [%s] Login Berhasil! User: %s (ID: %s)\n", now, email, user.ID)

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
