package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/muhammadfarrasfajri/filantropi/model"
	"github.com/muhammadfarrasfajri/filantropi/repository"
	"github.com/muhammadfarrasfajri/filantropi/utils"
)

type UserService struct {
	UserRepo  repository.UserRepo
	RegisRepo repository.RegisterRepo
}

func NewUserService(userService repository.UserRepo, regisRepo repository.RegisterRepo) *UserService {
	return &UserService{
		UserRepo:  userService,
		RegisRepo: regisRepo,
	}
}

func (s *UserService) FindById(id string) (*model.User, error) {
	now := time.Now().Format("2006-01-02 15:04:05")

	// 1. LOG: Proses Pencarian
	fmt.Printf("[USER-SERVICE] [%s] Mencari data user dengan ID: %s\n", now, id)

	resurl, err := s.UserRepo.FindDonorsById(id)
	if err != nil {
		// 2. LOG: Jika Error atau Data Tidak Ada
		fmt.Printf("[ERROR] [%s] Gagal mendapatkan user ID %s: %v\n", now, id, err)
		return nil, err
	}

	fmt.Println(resurl)

	// 3. LOG: Berhasil Ditemukan
	fmt.Printf("[SUCCESS] [%s] User ID %s ditemukan. Email: %s\n", now, id, resurl.Email)

	return resurl, nil
}

func (s *UserService) FindBeneficiaryById(userId string) (*model.User, *model.BeneficiaryProfile, error) {
	now := time.Now().Format("2006-01-02 15:04:05")

	// 1. LOG: Proses Pencarian
	fmt.Printf("[USER-SERVICE] [%s] Mencari data user dengan ID: %s\n", now, userId)

	user, profile, err := s.UserRepo.FindBeneficiaryById(userId)
	if err != nil {
		// 2. LOG: Jika Error atau Data Tidak Ada
		fmt.Printf("[ERROR] [%s] Gagal mendapatkan user ID %s: %v\n", now, userId, err)
		return nil, nil, err
	}
	fmt.Println(profile.UrlKTP)

	// 3. LOG: Berhasil Ditemukan
	fmt.Printf("[SUCCESS] [%s] User ID %s ditemukan. Email: %s\n", now, userId, user.Email)

	return user, profile, nil
}

func (s *UserService) UpdateDonors(userID string, walletAddress string, fullName string, photoProfile string) error {
	ctx := context.Background()

	if walletAddress != "" {
		if !strings.HasPrefix(walletAddress, "0x") {
			return errors.New("format wallet address tidak valid")
		}
	}

	// 3. Panggil Repository (yang sudah pakai Transaction tadi)
	err := s.UserRepo.UpdateDonors(ctx, userID, walletAddress, fullName, photoProfile)
	if err != nil {
		// Kamu bisa membungkus error agar lebih jelas asalnya
		return fmt.Errorf("service update donor: %w", err)
	}

	return nil
}

func (s *UserService) UpdateProfileBeneficiaries(ctx context.Context, userId string, profile model.BeneficiaryProfile) error {
	now := time.Now().Format("2006-01-02 15:04:05")

	if userId == "" {
		return errors.New("user id tidak boleh kosong")
	}

	// 1. Ambil data lama untuk pembanding
	existingUser, existingBeneficiary, err := s.UserRepo.FindBeneficiaryById(userId)
	if err != nil {
		return errors.New("user tidak ditemukan")
	}

	// 2. LOGIC WALLET (PENTING!)
	walletChanged := false // Flag penanda apakah wallet berubah

	if profile.WalletAddress != "" && profile.WalletAddress != existingUser.WalletAddress {
		// User menginput wallet BARU, tandai flag
		walletChanged = true

		isExistsWallet, err := s.RegisRepo.IsWalletAddressExists(profile.WalletAddress)
		if err != nil {
			return err
		}
		if isExistsWallet {
			return fmt.Errorf("wallet %s sudah digunakan oleh akun lain", profile.WalletAddress)
		}

		// Validasi format (Web3 Logic)
		if !strings.HasPrefix(profile.WalletAddress, "0x") || len(profile.WalletAddress) != 42 {
			return errors.New("format wallet address tidak valid")
		}
	} else {
		// Kalau user tidak input wallet, tetap pakai wallet yang lama
		profile.WalletAddress = existingUser.WalletAddress
	}

	// 3. Tentukan Tipe (Agar tidak kosong saat masuk ke Repo)
	if profile.BeneficiaryType == "" {
		profile.BeneficiaryType = existingUser.BeneficiaryType
	}

	if profile.PhotoProfile == "" {
		profile.PhotoProfile = existingBeneficiary.PhotoProfile
	}

	// 4. VALIDASI BUSINESS LOGIC
	if err := validateBeneficiaryFields(profile); err != nil {
		return err
	}

	fmt.Printf("[INFO] [%s] Processing update for: %s\n", now, profile.FullName)

	// 5. EKSEKUSI DATABASE
	err = s.UserRepo.UpdateProfileBeneficiary(ctx, userId, profile)
	if err != nil {
		fmt.Printf("[ERROR] [%s] Gagal update DB: %v\n", now, err)
		return fmt.Errorf("gagal memperbarui data: %v", err)
	}

	// 6. SINKRONISASI ALCHEMY (Menggunakan Goroutine & Logika Cerdas)
	// Hanya hubungi Alchemy JIKA wallet benar-benar diganti oleh user
	if walletChanged && profile.BeneficiaryType == "individual" {
		go func() {
			var errWebhook error

			// Cek apakah wallet sebelumnya kosong atau sudah ada isinya
			if existingUser.WalletAddress == "" {
				// Kasus A: Dulu kosong, sekarang diisi (Pakai Add)
				errWebhook = utils.AddWalletToAlchemyWebhook(profile.WalletAddress)
			} else {
				// Kasus B: Dulu ada, sekarang ditukar (Pakai Swap)
				errWebhook = utils.SwapWalletInAlchemyWebhook(existingUser.WalletAddress, profile.WalletAddress)
			}

			if errWebhook != nil {
				// Print error di terminal saja, jangan hentikan proses karena DB sudah terupdate
				fmt.Printf("[ERROR] [%s] Gagal sinkronisasi wallet ke Alchemy: %v\n", time.Now().Format("2006-01-02 15:04:05"), errWebhook)
			}
		}()
	}

	// 7. Kembalikan Sukses (Frontend tidak perlu menunggu proses Alchemy selesai)
	return nil
}

// Fungsi pembantu agar kodingan utama tidak kepanjangan
func validateBeneficiaryFields(p model.BeneficiaryProfile) error {
	if p.FullName == "" {
		return errors.New("nama lengkap wajib diisi")
	}

	if p.BeneficiaryType == "individual" {
		if p.Nik != "" && len(p.Nik) != 16 {
			return errors.New("NIK harus 16 digit")
		}
	} else if p.BeneficiaryType == "organization" {
		if (p.RegistrationNumber == nil || *p.RegistrationNumber == "") &&
			(p.Npwp == nil || *p.Npwp == "") {
			return errors.New("organisasi wajib mengisi No. Registrasi atau NPWP")
		}
	}
	return nil
}

func (s *UserService) PostWallet(wallet string) error {
	return utils.AddWalletToAlchemyWebhook(wallet)
}
