package service

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"net/smtp"
	"os"
)

//go:embed templates
var templateFS embed.FS

type EmailService struct {
	smtpHost string
	smtpPort string
	smtpUser string
	smtpPass string
	baseURL  string
}

func NewEmailService() *EmailService {
	return &EmailService{
		smtpHost: os.Getenv("SMTP_HOST"),
		smtpPort: os.Getenv("SMTP_PORT"),
		smtpUser: os.Getenv("SMTP_USER"),
		smtpPass: os.Getenv("SMTP_PASS"),
		baseURL:  os.Getenv("FRONTEND_URL"),
	}
}

func (s *EmailService) SendPasswordReset(toEmail, userName, token string) error {
	tmpl, err := template.ParseFS(templateFS, "templates/reset_password.html")
	if err != nil {
		return fmt.Errorf("parsing email template: %w", err)
	}

	link := fmt.Sprintf("%s/reset-password?token=%s", s.baseURL, token)
	data := struct {
		Name string
		Link string
	}{
		Name: userName,
		Link: link,
	}

	var body bytes.Buffer
	if err := tmpl.Execute(&body, data); err != nil {
		return fmt.Errorf("executing email template: %w", err)
	}

	headers := "MIME-version: 1.0;\n" +
		"Content-Type: text/html; charset=\"UTF-8\";\n" +
		fmt.Sprintf("From: PMS Support <%s>\n", s.smtpUser) +
		fmt.Sprintf("To: %s\n", toEmail) +
		"Subject: Recuperación de Contraseña\n\n"

	msg := []byte(headers + body.String())

	auth := smtp.PlainAuth("", s.smtpUser, s.smtpPass, s.smtpHost)
	addr := fmt.Sprintf("%s:%s", s.smtpHost, s.smtpPort)

	if err := smtp.SendMail(addr, auth, s.smtpUser, []string{toEmail}, msg); err != nil {
		return fmt.Errorf("sending email via smtp: %w", err)
	}

	return nil
}
