// internal/utils/email.go
package utils

import (
	"fmt"
	"net/smtp"
	"os"
)

func SendRejectionEmail(toEmail string, reason string) error {
	// Konfigurasi SMTP (Contoh menggunakan Gmail)
	// Penting: Gunakan "App Password" Gmail, bukan password akun biasa
	from := "farrasfajri@gmail.com"
	password := os.Getenv("PASSWORD_GOOGLE_SEND_EMAIL")
	smtpHost := "smtp.gmail.com"
	smtpPort := "587"

	// Isi Email
	subject := "Subject: Update Verifikasi Akun Filantropi\n"
	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"

	body := fmt.Sprintf(`
		<h3>Mohon Maaf, Verifikasi Anda Ditolak</h3>
		<p>Tim admin kami telah meninjau data Anda, namun kami belum bisa memverifikasi akun Anda karena alasan berikut:</p>
		<blockquote style="background: #f8d7da; padding: 10px; border-left: 5px solid #dc3545;">
			<strong>%s</strong>
		</blockquote>
		<p>Silakan login kembali ke aplikasi dan perbarui data profil Anda sesuai dengan catatan di atas.</p>
		<br>
		<p>Terima kasih,<br>Tim Koperasi Gerai / Filantropi</p>
	`, reason)

	msg := []byte(subject + mime + body)

	// Proses pengiriman
	auth := smtp.PlainAuth("", from, password, smtpHost)
	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, []string{toEmail}, msg)
	return err
}
