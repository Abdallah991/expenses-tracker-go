package email

import (
	"fmt"
	"os"

	"github.com/resend/resend-go/v2"
)

// EmailService handles sending emails via Resend
type EmailService struct {
	client    *resend.Client
	fromEmail string
	appURL    string
}

// NewEmailService creates a new email service instance
func NewEmailService() (*EmailService, error) {
	apiKey := os.Getenv("RESEND_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("RESEND_API_KEY environment variable is required")
	}

	fromEmail := os.Getenv("FROM_EMAIL")
	if fromEmail == "" {
		return nil, fmt.Errorf("FROM_EMAIL environment variable is required")
	}

	appURL := os.Getenv("APP_URL")
	if appURL == "" {
		appURL = "http://localhost:8080"
	}

	client := resend.NewClient(apiKey)

	return &EmailService{
		client:    client,
		fromEmail: fromEmail,
		appURL:    appURL,
	}, nil
}

// SendVerificationEmail sends an email verification email
func (es *EmailService) SendVerificationEmail(to, token string) error {
	verificationURL := fmt.Sprintf("%s/auth/verify-email?token=%s", es.appURL, token)

	subject := "Verify Your Email - Expenses Tracker"
	htmlContent := fmt.Sprintf(`
		<!DOCTYPE html>
		<html>
		<head>
			<meta charset="utf-8">
			<title>Verify Your Email</title>
			<style>
				body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
				.container { max-width: 600px; margin: 0 auto; padding: 20px; }
				.header { background-color: #4CAF50; color: white; padding: 20px; text-align: center; }
				.content { padding: 20px; background-color: #f9f9f9; }
				.button { display: inline-block; padding: 12px 24px; background-color: #4CAF50; color: white; text-decoration: none; border-radius: 4px; margin: 20px 0; }
				.footer { padding: 20px; text-align: center; color: #666; font-size: 12px; }
			</style>
		</head>
		<body>
			<div class="container">
				<div class="header">
					<h1>Welcome to Expenses Tracker!</h1>
				</div>
				<div class="content">
					<h2>Verify Your Email Address</h2>
					<p>Thank you for registering with Expenses Tracker. To complete your registration and start tracking your expenses, please verify your email address by clicking the button below:</p>
					
					<div style="text-align: center;">
						<a href="%s" class="button">Verify Email Address</a>
					</div>
					
					<p>If the button doesn't work, you can copy and paste this link into your browser:</p>
					<p style="word-break: break-all; background-color: #eee; padding: 10px; border-radius: 4px;">%s</p>
					
					<p><strong>Important:</strong> This verification link will expire in 24 hours for security reasons.</p>
					
					<p>If you didn't create an account with Expenses Tracker, you can safely ignore this email.</p>
				</div>
				<div class="footer">
					<p>This email was sent from Expenses Tracker. Please do not reply to this email.</p>
				</div>
			</div>
		</body>
		</html>
	`, verificationURL, verificationURL)

	textContent := fmt.Sprintf(`
		Welcome to Expenses Tracker!
		
		Thank you for registering with Expenses Tracker. To complete your registration and start tracking your expenses, please verify your email address by visiting this link:
		
		%s
		
		Important: This verification link will expire in 24 hours for security reasons.
		
		If you didn't create an account with Expenses Tracker, you can safely ignore this email.
		
		This email was sent from Expenses Tracker. Please do not reply to this email.
	`, verificationURL)

	params := &resend.SendEmailRequest{
		From:    es.fromEmail,
		To:      []string{to},
		Subject: subject,
		Html:    htmlContent,
		Text:    textContent,
	}

	_, err := es.client.Emails.Send(params)
	return err
}

// SendPasswordResetEmail sends a password reset email
func (es *EmailService) SendPasswordResetEmail(to, token string) error {
	resetURL := fmt.Sprintf("%s/auth/reset-password?token=%s", es.appURL, token)

	subject := "Reset Your Password - Expenses Tracker"
	htmlContent := fmt.Sprintf(`
		<!DOCTYPE html>
		<html>
		<head>
			<meta charset="utf-8">
			<title>Reset Your Password</title>
			<style>
				body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
				.container { max-width: 600px; margin: 0 auto; padding: 20px; }
				.header { background-color: #f44336; color: white; padding: 20px; text-align: center; }
				.content { padding: 20px; background-color: #f9f9f9; }
				.button { display: inline-block; padding: 12px 24px; background-color: #f44336; color: white; text-decoration: none; border-radius: 4px; margin: 20px 0; }
				.footer { padding: 20px; text-align: center; color: #666; font-size: 12px; }
				.warning { background-color: #fff3cd; border: 1px solid #ffeaa7; padding: 15px; border-radius: 4px; margin: 20px 0; }
			</style>
		</head>
		<body>
			<div class="container">
				<div class="header">
					<h1>Password Reset Request</h1>
				</div>
				<div class="content">
					<h2>Reset Your Password</h2>
					<p>We received a request to reset your password for your Expenses Tracker account. If you made this request, click the button below to reset your password:</p>
					
					<div style="text-align: center;">
						<a href="%s" class="button">Reset Password</a>
					</div>
					
					<p>If the button doesn't work, you can copy and paste this link into your browser:</p>
					<p style="word-break: break-all; background-color: #eee; padding: 10px; border-radius: 4px;">%s</p>
					
					<div class="warning">
						<p><strong>Security Notice:</strong></p>
						<ul>
							<li>This password reset link will expire in 1 hour for security reasons.</li>
							<li>If you didn't request a password reset, please ignore this email.</li>
							<li>Your password will remain unchanged until you click the link above.</li>
						</ul>
					</div>
					
					<p>For your security, if you didn't request this password reset, please contact our support team immediately.</p>
				</div>
				<div class="footer">
					<p>This email was sent from Expenses Tracker. Please do not reply to this email.</p>
				</div>
			</div>
		</body>
		</html>
	`, resetURL, resetURL)

	textContent := fmt.Sprintf(`
		Password Reset Request
		
		We received a request to reset your password for your Expenses Tracker account. If you made this request, visit this link to reset your password:
		
		%s
		
		Security Notice:
		- This password reset link will expire in 1 hour for security reasons.
		- If you didn't request a password reset, please ignore this email.
		- Your password will remain unchanged until you click the link above.
		
		For your security, if you didn't request this password reset, please contact our support team immediately.
		
		This email was sent from Expenses Tracker. Please do not reply to this email.
	`, resetURL)

	params := &resend.SendEmailRequest{
		From:    es.fromEmail,
		To:      []string{to},
		Subject: subject,
		Html:    htmlContent,
		Text:    textContent,
	}

	_, err := es.client.Emails.Send(params)
	return err
}
