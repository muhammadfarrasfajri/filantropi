package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

type AlchemyWebhookUpdatePayload struct {
	WebhookID         string   `json:"webhook_id"`
	AddressesToAdd    []string `json:"addresses_to_add"`
	AddressesToRemove []string `json:"addresses_to_remove"`
}

// Fungsi untuk menembak API Alchemy
func AddWalletToAlchemyWebhook(newWalletAddress string) error {
	// =======================================================
	// WAJIB DIISI: Ambil dari Dashboard Alchemy kamu
	// =======================================================
	webhookID := os.Getenv("ALCHEMY_WEBHOOK_ID")        // ID Webhook kamu
	alchemyAuthToken := os.Getenv("ALCHEMY_AUTH_TOKEN") // API Key / Auth Token Alchemy

	url := "https://dashboard.alchemy.com/api/update-webhook-addresses"

	// Siapkan data JSON (hanya mengisi AddressesToAdd)
	payload := AlchemyWebhookUpdatePayload{
		WebhookID:         webhookID,
		AddressesToAdd:    []string{newWalletAddress},
		AddressesToRemove: []string{},
	}

	jsonValue, _ := json.Marshal(payload)

	// Buat request HTTP PATCH
	req, err := http.NewRequest("PATCH", url, bytes.NewBuffer(jsonValue))
	if err != nil {
		return err
	}

	// Atur Header agar diizinkan oleh Alchemy
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Alchemy-Token", alchemyAuthToken)

	// Eksekusi Request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// BAGIAN YANG DIPERBAIKI: Tangkap pesan error asli dari Alchemy
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body) // Baca isi balasan Alchemy
		return fmt.Errorf("gagal add webhook Alchemy, Status: %d, Response: %s", resp.StatusCode, string(bodyBytes))
	}

	// Saran: Gunakan log bawaan Golang (log.Printf) agar ada keterangan waktunya
	fmt.Printf("[SYSTEM] Wallet %s otomatis ditambahkan ke pantauan Alchemy!\n", newWalletAddress)
	return nil
}

func SwapWalletInAlchemyWebhook(oldWallet string, newWallet string) error {

	// Ambil token dari file .env agar aman
	webhookID := os.Getenv("ALCHEMY_WEBHOOK_ID")
	alchemyAuthToken := os.Getenv("ALCHEMY_AUTH_TOKEN")

	url := "https://dashboard.alchemy.com/api/update-webhook-addresses"

	// Masukkan dompet baru ke Adds, dan dompet lama ke Removes
	payload := AlchemyWebhookUpdatePayload{
		WebhookID:         webhookID,
		AddressesToAdd:    []string{newWallet},
		AddressesToRemove: []string{oldWallet},
	}

	jsonValue, _ := json.Marshal(payload)

	req, err := http.NewRequest("PATCH", url, bytes.NewBuffer(jsonValue))
	if err != nil {
		return err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Alchemy-Token", alchemyAuthToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// BAGIAN YANG DIPERBAIKI: Tangkap pesan error asli dari Alchemy
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body) // Baca isi balasan Alchemy
		return fmt.Errorf("gagal swap webhook Alchemy, Status: %d, Response: %s", resp.StatusCode, string(bodyBytes))
	}

	fmt.Printf("[SYSTEM] Alchemy Swap Sukses: %s dihapus, %s ditambahkan!\n", oldWallet, newWallet)
	return nil
}

func RemoveWalletFromAlchemyWebhook(walletToRemove string) error {

	// Ambil token dari file .env
	webhookID := os.Getenv("ALCHEMY_WEBHOOK_ID")
	alchemyAuthToken := os.Getenv("ALCHEMY_AUTH_TOKEN")

	url := "https://dashboard.alchemy.com/api/update-webhook-addresses"

	// Siapkan data JSON: Adds dikosongkan, Removes diisi
	payload := AlchemyWebhookUpdatePayload{
		WebhookID:         webhookID,
		AddressesToAdd:    []string{},               // Kosong
		AddressesToRemove: []string{walletToRemove}, // Isi dengan dompet yang mau dihapus
	}

	jsonValue, _ := json.Marshal(payload)

	// Buat request HTTP PATCH
	req, err := http.NewRequest("PATCH", url, bytes.NewBuffer(jsonValue))
	if err != nil {
		return err
	}

	// Atur Header agar diizinkan oleh Alchemy
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Alchemy-Token", alchemyAuthToken)

	// Eksekusi Request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Tangkap pesan error asli jika Alchemy menolak
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("gagal menghapus webhook Alchemy, Status: %d, Response: %s", resp.StatusCode, string(bodyBytes))
	}

	// Print log sukses di terminal
	fmt.Printf("[SYSTEM] Wallet %s otomatis DIHAPUS dari pantauan Alchemy!\n", walletToRemove)
	return nil
}
