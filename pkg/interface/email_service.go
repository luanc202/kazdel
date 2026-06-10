package interfaces

type EmailService interface {
	SendEmail(to []string, subject string, body string) error
	SendVerificationEmail(to, username, verificationLink string) error
	SendPasswordResetEmail(to, username, resetLink string) error
	SendReportEmail(reportedURL, reason, description, reporterEmail string) error
}
