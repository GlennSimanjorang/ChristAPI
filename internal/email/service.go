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
	m.SetHeader("Subject", "ChristAPI - Email Verification Code")

	m.SetBody("text/html", fmt.Sprintf(`
	<!DOCTYPE html>
	<html>
	<head>
	<meta charset="UTF-8">
	<title>ChristAPI Verification</title>
	</head>

	<body style="margin:0;padding:0;background:#f5f5f5;font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,Helvetica,Arial,sans-serif;color:#111827;">

	<table role="presentation" width="100%%" cellspacing="0" cellpadding="0">
	<tr>
	<td align="center" style="padding:48px 16px;">

	<table role="presentation" width="560" cellspacing="0" cellpadding="0"
	style="background:#ffffff;border:1px solid #e5e7eb;border-radius:12px;overflow:hidden;">

	<tr>
	<td style="padding:32px 40px;border-bottom:1px solid #e5e7eb;">
		<h2 style="margin:0;font-size:22px;font-weight:600;color:#111827;">
			Verify your email
		</h2>
	</td>
	</tr>

	<tr>
	<td style="padding:32px 40px;line-height:1.7;font-size:15px;color:#374151;">

	<p style="margin:0 0 16px;">
	Use the verification code below to complete your email verification.
	</p>

	<div style="
		margin:32px 0;
		padding:18px;
		text-align:center;
		border:1px solid #d1d5db;
		border-radius:8px;
		background:#f9fafb;
	">
		<span style="
			font-size:34px;
			font-weight:700;
			letter-spacing:8px;
			color:#111827;
			font-family:Consolas,Menlo,Monaco,monospace;
		">
			%s
		</span>
	</div>

	<p style="margin:0 0 12px;">
	This code will expire in <strong>5 minutes</strong>.
	</p>

	<p style="margin:0;">
	If you did not request this verification, you can safely ignore this email.
	</p>

	</td>
	</tr>

	<tr>
	<td style="
	padding:20px 40px;
	background:#f9fafb;
	border-top:1px solid #e5e7eb;
	font-size:13px;
	color:#6b7280;
	">
	This email was sent automatically by ChristAPI. Please do not reply to this message.
	</td>
	</tr>

	</table>

	</td>
	</tr>
	</table>

	</body>
	</html>
	`, otpCode))
	d := gomail.NewDialer(host, port, user, password)
	return d.DialAndSend(m)
}
