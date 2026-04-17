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

func (s *RegisterService) RegisterGoogle(user model.User) (map[string]interface{}, error) {
	ctx := context.Background()

	// Verification token id with Firebase Auth
	token, err := s.FirebaseAuth.VerifyIDToken(ctx, user.IdToken)
	if err != nil {
		return nil, err
	}

	// Check if wallet address already exists in the database
	if user.WalletAddress != "" {
		isExistsWallet, err := s.RegisRepo.IsWalletAddressExists(user.WalletAddress)
		if err != nil {
			return nil, err
		}
		if isExistsWallet {
			return nil, fmt.Errorf("wallet address already exists")
		}
	}

	// Generate user ID if not provided
	if user.ID == "" {
		user.ID = uuid.New().String()
	}

	// Extract email and name from token claims if not provided in the user struct
	if email, ok := token.Claims["email"].(string); ok {
		user.Email = email
	}
	// Extract name from token claims if not provided in the user struct
	if user.Name == "" {
		if n, ok := token.Claims["name"].(string); ok {
			user.Name = n
		}
	}

	// Check if email already exists in the database
	isExistsEmail, err := s.RegisRepo.IsEmailExists(user.Email)
	if err != nil {
		return nil, err
	}

	if isExistsEmail {
		return nil, fmt.Errorf("email already exists")
	}

	// Set default role if not provided
	if user.Role == "" {
		user.Role = "user"
	}

	// Generate access token
	accessToken, err := s.JWTManager.GenerateAccessToken(user.ID, user.Email, user.Role)
	if err != nil {
		return nil, err
	}

	// Generate refresh token
	refreshToken, err := s.JWTManager.GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, err
	}

	refreshTokenHash := middleware.HashToken(refreshToken)
	expiresAt := time.Now().Add(7 * 24 * time.Hour)

	// Create refresh token data
	tokenData := model.RefreshToken{
		UserID:    user.ID,
		Token:     refreshTokenHash,
		ExpiresAt: expiresAt,
	}

	// Create user in the database
	err = s.RegisRepo.CreateUser(user)
	if err != nil {
		return nil, err
	}

	// Store refresh token in the database
	err = s.RefreshRepo.UpsertRefreshToken(tokenData)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"access_token": accessToken,
		"token_hash":   refreshToken,
		"user": map[string]interface{}{
			"id":    user.ID,
			"name":  user.Name,
			"email": user.Email,
			"role":  user.Role,
		},
	}, nil
}
