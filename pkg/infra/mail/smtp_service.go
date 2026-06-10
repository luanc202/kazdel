package mail

import (
	"crypto/tls"
	"fmt"
	"kazdel/pkg/infra/config"
	"log/slog"
	"net/smtp"
	"strings"
)

type SMTPMailService struct {
	env *config.EnvConfig
}

func NewSMTPMailService() *SMTPMailService {
	return &SMTPMailService{
		env: config.GetEnvConfig(),
	}
}

// sendAsync handles the actual sending of the email in a background goroutine
func (s *SMTPMailService) sendAsync(to []string, subject, body string) {
	go func() {
		if !s.env.MAIL_ENABLED {
			slog.Info("Mail is disabled. Skipping email sending", "to", to, "subject", subject)
			return
		}

		headers := make(map[string]string)
		headers["From"] = s.env.MAIL_FROM
		headers["To"] = strings.Join(to, ",")
		headers["Subject"] = subject
		headers["MIME-Version"] = "1.0"
		headers["Content-Type"] = "text/html; charset=\"utf-8\""

		message := ""
		for k, v := range headers {
			message += fmt.Sprintf("%s: %s\r\n", k, v)
		}
		message += "\r\n" + body

		var auth smtp.Auth
		if s.env.MAIL_USER != "" {
			auth = smtp.PlainAuth("", s.env.MAIL_USER, s.env.MAIL_PASSWORD, s.env.MAIL_HOST)
		}
		var err error
		addr := fmt.Sprintf("%s:%s", s.env.MAIL_HOST, s.env.MAIL_PORT)

		if s.env.MAIL_SECURE {
			// TLS configuration
			tlsconfig := &tls.Config{
				InsecureSkipVerify: true,
				ServerName:         s.env.MAIL_HOST,
			}

			// Dial TCP connection
			conn, err := tls.Dial("tcp", addr, tlsconfig)
			if err != nil {
				slog.Error("Failed to dial SMTP server", "error", err)
				return
			}
			defer conn.Close()

			client, err := smtp.NewClient(conn, s.env.MAIL_HOST)
			if err != nil {
				slog.Error("Failed to create SMTP client", "error", err)
				return
			}

			if auth != nil {
				if err = client.Auth(auth); err != nil {
					slog.Error("Failed to authenticate with SMTP server", "error", err)
					return
				}
			}

			if err = client.Mail(s.env.MAIL_FROM); err != nil {
				slog.Error("Failed to set MAIL FROM", "error", err)
				return
			}

			for _, rcpt := range to {
				if err = client.Rcpt(rcpt); err != nil {
					slog.Error("Failed to set RCPT TO", "error", err)
					return
				}
			}

			w, err := client.Data()
			if err != nil {
				slog.Error("Failed to open Data writer", "error", err)
				return
			}

			_, err = w.Write([]byte(message))
			if err != nil {
				slog.Error("Failed to write message body", "error", err)
				return
			}

			err = w.Close()
			if err != nil {
				slog.Error("Failed to close Data writer", "error", err)
				return
			}

			client.Quit()

		} else {
			// Plain SMTP
			err = smtp.SendMail(addr, auth, s.env.MAIL_FROM, to, []byte(message))
		}

		if err != nil {
			slog.Error("Failed to send email", "to", to, "subject", subject, "error", err)
		} else {
			slog.Info("Email sent successfully", "to", to, "subject", subject)
		}
	}()
}

func (s *SMTPMailService) SendEmail(to []string, subject string, body string) error {
	s.sendAsync(to, subject, body)
	return nil
}

func (s *SMTPMailService) SendVerificationEmail(to, username, verificationLink string) error {
	subject := "Verify your email address"
	body := fmt.Sprintf(`
		<h1>Welcome to Kazdel, %s!</h1>
		<p>Please click the link below to verify your email address:</p>
		<p><a href="%s">%s</a></p>
	`, username, verificationLink, verificationLink)
	
	s.sendAsync([]string{to}, subject, body)
	return nil
}

func (s *SMTPMailService) SendPasswordResetEmail(to, username, resetLink string) error {
	subject := "Reset your password"
	body := fmt.Sprintf(`
		<h1>Hello %s,</h1>
		<p>You requested a password reset. Please click the link below to set a new password:</p>
		<p><a href="%s">%s</a></p>
		<p>If you did not request this, please ignore this email.</p>
	`, username, resetLink, resetLink)

	s.sendAsync([]string{to}, subject, body)
	return nil
}

func (s *SMTPMailService) SendReportEmail(reportedURL, reason, description, reporterEmail string) error {
	subject := "New Malicious URL Report"
	body := fmt.Sprintf(`
		<h1>New Malicious URL Report</h1>
		<p><strong>Reported URL:</strong> %s</p>
		<p><strong>Reason:</strong> %s</p>
		<p><strong>Description:</strong> %s</p>
		<p><strong>Reporter Email:</strong> %s</p>
	`, reportedURL, reason, description, reporterEmail)

	// Here it sends to the admin's email. We'll use MAIL_FROM as the admin's destination for now.
	s.sendAsync([]string{s.env.MAIL_FROM}, subject, body)
	return nil
}
