package email

import (
	"fmt"
	"os"
	"strconv"

	"gopkg.in/gomail.v2"
)

func SendOTP(email, otpCode string) error {
	host := os.Getenv("SMTP_HOST")
	portStr := os.Getenv("SMTP_PORT")
	user := os.Getenv("SMTP_USER")
	password := os.Getenv("SMTP_PASSWORD")
	sender := os.Getenv("SENDER_EMAIL")

	if host == "" || user == "" || password == "" || sender == "" {
		return fmt.Errorf("SMTP configuration incomplete")
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return fmt.Errorf("invalid SMTP_PORT: %w", err)
	}

	m := gomail.NewMessage()
	m.SetHeader("From", sender)
	m.SetHeader("To", email)
	m.SetHeader("Subject", "ChristAPI - Kode Verifikasi OTP")
	m.SetBody("text/html", fmt.Sprintf(`
		<h2>Verifikasi Email Anda</h2>
		<p>Berikut adalah kode OTP untuk verifikasi akun Anda:</p>
		<h1 style="font-size: 32px; letter-spacing: 2px; color: #3498db;">%s</h1>
		<p>Kode ini berlaku selama <strong>5 menit</strong>.</p>
		<p>Jika Anda tidak melakukan permintaan ini, abaikan email ini.</p>
		<hr>
		<p style="font-size: 12px; color: #7f8c8d;">ChristAPI Team</p>
	`, otpCode))

	d := gomail.NewDialer(host, port, user, password)
	return d.DialAndSend(m)
}
