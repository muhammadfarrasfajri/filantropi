package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

const alchemyURL = "https://polygon-mainnet.g.alchemy.com/v2/EoACZFbhYDxPu8TGpit7u"

func VerifyDonationOnChain(txHash string, receiverWallet string, expectedAmount float64) (bool, error) {
	// 1. Request Detail Transaksi (Untuk cek Nominal & Penerima)
	// ... (Kode kamu yang eth_getTransactionByHash tetap di sini) ...

	// 2. Request Transaction Receipt (Wajib untuk cek apakah sukses/gagal)
	payloadReceipt := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "eth_getTransactionReceipt",
		"params":  []string{txHash},
		"id":      1,
	}

	bodyReceipt, _ := json.Marshal(payloadReceipt)
	respReceipt, err := http.Post(alchemyURL, "application/json", bytes.NewBuffer(bodyReceipt))
	if err != nil {
		return false, fmt.Errorf("gagal cek receipt ke Alchemy")
	}
	defer respReceipt.Body.Close()

	var receiptResult map[string]interface{}
	json.NewDecoder(respReceipt.Body).Decode(&receiptResult)

	receiptData, ok := receiptResult["result"].(map[string]interface{})
	if !ok || receiptData == nil {
		return false, fmt.Errorf("receipt belum tersedia (transaksi mungkin masih pending)")
	}

	// --- PENGECEKAN KEAMANAN TAMBAHAN ---

	// C. Cek Status Transaksi (0x1 = Success, 0x0 = Failed)
	status := fmt.Sprintf("%v", receiptData["status"])
	if status != "0x1" {
		return false, fmt.Errorf("transaksi gagal di blockchain (reverted)")
	}

	// D. Cek Konfirmasi (Optional tapi disarankan)
	// Memastikan transaksi sudah masuk ke dalam blok (bukan lagi pending)
	if receiptData["blockNumber"] == nil {
		return false, fmt.Errorf("transaksi belum dikonfirmasi oleh blok")
	}

	return true, nil
}
