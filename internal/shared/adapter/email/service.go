package email

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"net/smtp"
	"os"
	"time"
)

var templateFS embed.FS

type Service struct {
	smtpHost string
	smtpPort string
	smtpUser string
	smtpPass string
	baseURL  string
}

func NewService() *Service {
	return &Service{
		smtpHost: os.Getenv("SMTP_HOST"),
		smtpPort: os.Getenv("SMTP_PORT"),
		smtpUser: os.Getenv("SMTP_USER"),
		smtpPass: os.Getenv("SMTP_PASS"),
		baseURL:  os.Getenv("FRONTEND_URL"),
	}
}

func (s *Service) SendPasswordReset(toEmail, userName, token string) error {
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
	return s.sendEmail(toEmail, "Password Reset", tmpl, data)
}

func (s *Service) SendReservationConfirmed(toEmail, guestName, reservationCode string, checkIn, checkOut time.Time) error {
	tmpl, err := template.ParseFS(templateFS, "templates/reservation_confirmed.html")
	if err != nil {
		return fmt.Errorf("parsing email template: %w", err)
	}
	data := struct {
		GuestName        string
		ReservationCode  string
		CheckInDate      string
		CheckOutDate     string
	}{
		GuestName:       guestName,
		ReservationCode: reservationCode,
		CheckInDate:      checkIn.Format("Jan 2, 2006"),
		CheckOutDate:     checkOut.Format("Jan 2, 2006"),
	}
	return s.sendEmail(toEmail, "Reservation Confirmed", tmpl, data)
}

func (s *Service) SendReservationCancelled(toEmail, guestName, reservationCode string) error {
	tmpl, err := template.ParseFS(templateFS, "templates/reservation_cancelled.html")
	if err != nil {
		return fmt.Errorf("parsing email template: %w", err)
	}
	data := struct {
		GuestName        string
		ReservationCode  string
	}{
		GuestName:       guestName,
		ReservationCode: reservationCode,
	}
	return s.sendEmail(toEmail, "Reservation Cancelled", tmpl, data)
}

func (s *Service) sendEmail(to, subject string, tmpl *template.Template, data interface{}) error {
	var body bytes.Buffer
	if err := tmpl.Execute(&body, data); err != nil {
		return fmt.Errorf("executing email template: %w", err)
	}
	headers := "MIME-version: 1.0;\n" +
		"Content-Type: text/html; charset=\"UTF-8\";\n" +
		fmt.Sprintf("From: PMS Support <%s>\n", s.smtpUser) +
		fmt.Sprintf("To: %s\n", to) +
		fmt.Sprintf("Subject: %s\n\n", subject)
	msg := []byte(headers + body.String())
	auth := smtp.PlainAuth("", s.smtpUser, s.smtpPass, s.smtpHost)
	addr := fmt.Sprintf("%s:%s", s.smtpHost, s.smtpPort)
	if err := smtp.SendMail(addr, auth, s.smtpUser, []string{to}, msg); err != nil {
		return fmt.Errorf("sending email via smtp: %w", err)
	}
	return nil
}
