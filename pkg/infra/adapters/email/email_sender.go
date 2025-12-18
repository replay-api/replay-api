// Package email provides email sending adapters for authentication flows.
package email

import (
	"context"
	"fmt"
	"log/slog"
	"net/smtp"
	"strings"

	auth_out "github.com/replay-api/replay-api/pkg/domain/auth/ports/out"
)

// SMTPConfig holds SMTP server configuration
type SMTPConfig struct {
	Host        string
	Port        int
	Username    string
	Password    string
	FromEmail   string
	FromName    string
	AppName     string
	AppURL      string
}

// SMTPEmailSender implements EmailSender using SMTP
type SMTPEmailSender struct {
	config SMTPConfig
}

// NewSMTPEmailSender creates a new SMTP email sender
func NewSMTPEmailSender(config SMTPConfig) auth_out.EmailSender {
	return &SMTPEmailSender{config: config}
}

// SendVerificationEmail sends a verification email
func (s *SMTPEmailSender) SendVerificationEmail(ctx context.Context, email, token, code string, expiresAt string) error {
	verifyURL := fmt.Sprintf("%s/verify-email?token=%s", s.config.AppURL, token)

	subject := fmt.Sprintf("Verify your email for %s", s.config.AppName)
	body := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; line-height: 1.6; color: #34445C; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: linear-gradient(135deg, #FF4654 0%%, #FFC700 100%%); padding: 30px; text-align: center; }
        .header h1 { color: #F5F0E1; margin: 0; font-size: 28px; }
        .content { padding: 30px; background: #ffffff; }
        .code { font-size: 32px; font-weight: bold; letter-spacing: 8px; color: #FF4654; text-align: center; margin: 20px 0; }
        .button { display: inline-block; background: linear-gradient(135deg, #FF4654 0%%, #FFC700 100%%); color: #F5F0E1; padding: 15px 30px; text-decoration: none; font-weight: bold; margin: 20px 0; }
        .footer { text-align: center; padding: 20px; color: #888; font-size: 12px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🎮 %s</h1>
        </div>
        <div class="content">
            <h2>Verify your email address</h2>
            <p>Thanks for signing up! Please verify your email address by using the code below or clicking the button.</p>
            
            <div class="code">%s</div>
            
            <p style="text-align: center;">
                <a href="%s" class="button">Verify Email</a>
            </p>
            
            <p>This verification code expires at %s.</p>
            
            <p>If you didn't create an account, you can safely ignore this email.</p>
        </div>
        <div class="footer">
            <p>© %s. All rights reserved.</p>
        </div>
    </div>
</body>
</html>
`, s.config.AppName, code, verifyURL, expiresAt, s.config.AppName)

	return s.sendEmail(ctx, email, subject, body)
}

// SendMFACode sends an MFA code email
func (s *SMTPEmailSender) SendMFACode(ctx context.Context, email, code string, expiresAt string) error {
	subject := fmt.Sprintf("Your verification code for %s", s.config.AppName)
	body := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; line-height: 1.6; color: #34445C; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: linear-gradient(135deg, #FF4654 0%%, #FFC700 100%%); padding: 30px; text-align: center; }
        .header h1 { color: #F5F0E1; margin: 0; font-size: 28px; }
        .content { padding: 30px; background: #ffffff; }
        .code { font-size: 36px; font-weight: bold; letter-spacing: 10px; color: #FF4654; text-align: center; margin: 30px 0; background: #f8f8f8; padding: 20px; }
        .footer { text-align: center; padding: 20px; color: #888; font-size: 12px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🔐 %s</h1>
        </div>
        <div class="content">
            <h2>Your verification code</h2>
            <p>Use this code to complete your sign-in:</p>
            
            <div class="code">%s</div>
            
            <p><strong>This code expires at %s.</strong></p>
            
            <p>If you didn't request this code, please change your password immediately and contact support.</p>
        </div>
        <div class="footer">
            <p>© %s. All rights reserved.</p>
        </div>
    </div>
</body>
</html>
`, s.config.AppName, code, expiresAt, s.config.AppName)

	return s.sendEmail(ctx, email, subject, body)
}

// SendPasswordResetEmail sends a password reset email
func (s *SMTPEmailSender) SendPasswordResetEmail(ctx context.Context, email, token string, expiresAt string) error {
	resetURL := fmt.Sprintf("%s/reset-password?token=%s", s.config.AppURL, token)

	subject := fmt.Sprintf("Reset your password for %s", s.config.AppName)
	body := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; line-height: 1.6; color: #34445C; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: linear-gradient(135deg, #FF4654 0%%, #FFC700 100%%); padding: 30px; text-align: center; }
        .header h1 { color: #F5F0E1; margin: 0; font-size: 28px; }
        .content { padding: 30px; background: #ffffff; }
        .button { display: inline-block; background: linear-gradient(135deg, #FF4654 0%%, #FFC700 100%%); color: #F5F0E1; padding: 15px 30px; text-decoration: none; font-weight: bold; margin: 20px 0; }
        .warning { background: #fff3cd; border: 1px solid #ffc107; padding: 15px; margin: 20px 0; }
        .footer { text-align: center; padding: 20px; color: #888; font-size: 12px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🔑 %s</h1>
        </div>
        <div class="content">
            <h2>Reset your password</h2>
            <p>We received a request to reset your password. Click the button below to create a new password:</p>
            
            <p style="text-align: center;">
                <a href="%s" class="button">Reset Password</a>
            </p>
            
            <p>This link expires at %s.</p>
            
            <div class="warning">
                <strong>⚠️ Security Notice:</strong> If you didn't request a password reset, please ignore this email. Your password won't be changed.
            </div>
        </div>
        <div class="footer">
            <p>© %s. All rights reserved.</p>
        </div>
    </div>
</body>
</html>
`, s.config.AppName, resetURL, expiresAt, s.config.AppName)

	return s.sendEmail(ctx, email, subject, body)
}

// sendEmail sends an email using SMTP
func (s *SMTPEmailSender) sendEmail(ctx context.Context, to, subject, body string) error {
	from := s.config.FromEmail
	if s.config.FromName != "" {
		from = fmt.Sprintf("%s <%s>", s.config.FromName, s.config.FromEmail)
	}

	headers := []string{
		fmt.Sprintf("From: %s", from),
		fmt.Sprintf("To: %s", to),
		fmt.Sprintf("Subject: %s", subject),
		"MIME-Version: 1.0",
		"Content-Type: text/html; charset=UTF-8",
	}

	message := []byte(strings.Join(headers, "\r\n") + "\r\n\r\n" + body)

	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)
	auth := smtp.PlainAuth("", s.config.Username, s.config.Password, s.config.Host)

	err := smtp.SendMail(addr, auth, s.config.FromEmail, []string{to}, message)
	if err != nil {
		slog.ErrorContext(ctx, "failed to send email", "error", err, "to", to)
		return fmt.Errorf("failed to send email: %w", err)
	}

	slog.InfoContext(ctx, "email sent successfully", "to", to, "subject", subject)
	return nil
}

// NoopEmailSender is a no-operation email sender for testing/development
type NoopEmailSender struct {
	LogEmails bool
}

// NewNoopEmailSender creates a new no-op email sender
func NewNoopEmailSender(logEmails bool) auth_out.EmailSender {
	return &NoopEmailSender{LogEmails: logEmails}
}

// SendVerificationEmail logs the verification email instead of sending
func (s *NoopEmailSender) SendVerificationEmail(ctx context.Context, email, token, code string, expiresAt string) error {
	if s.LogEmails {
		slog.InfoContext(ctx, "📧 [NOOP] Verification email",
			"to", email,
			"token", token,
			"code", code,
			"expires_at", expiresAt)
	}
	return nil
}

// SendMFACode logs the MFA code instead of sending
func (s *NoopEmailSender) SendMFACode(ctx context.Context, email, code string, expiresAt string) error {
	if s.LogEmails {
		slog.InfoContext(ctx, "📧 [NOOP] MFA code email",
			"to", email,
			"code", code,
			"expires_at", expiresAt)
	}
	return nil
}

// SendPasswordResetEmail logs the password reset email instead of sending
func (s *NoopEmailSender) SendPasswordResetEmail(ctx context.Context, email, token string, expiresAt string) error {
	if s.LogEmails {
		slog.InfoContext(ctx, "📧 [NOOP] Password reset email",
			"to", email,
			"token", token,
			"expires_at", expiresAt)
	}
	return nil
}

// Ensure implementations satisfy the interface
var (
	_ auth_out.EmailSender = (*SMTPEmailSender)(nil)
	_ auth_out.EmailSender = (*NoopEmailSender)(nil)
)

