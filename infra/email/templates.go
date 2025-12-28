package email

import "fmt"

// GenerateVerificationEmailHTML generates the HTML content for email verification
func GenerateVerificationEmailHTML(verificationURL string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<title>Verify Your Email</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333; max-width: 600px; margin: 0 auto; padding: 20px;">
	<div style="background-color: #f4f4f4; padding: 20px; border-radius: 5px;">
		<h1 style="color: #2c3e50; margin-top: 0;">Verify Your Email</h1>
		<p>Thank you for registering with Anki Backend!</p>
		<p>Please click the button below to verify your email address:</p>
		<div style="text-align: center; margin: 30px 0;">
			<a href="%s" style="background-color: #3498db; color: white; padding: 12px 30px; text-decoration: none; border-radius: 5px; display: inline-block; font-weight: bold;">Verify Email</a>
		</div>
		<p>Or copy and paste this link into your browser:</p>
		<p style="word-break: break-all; color: #7f8c8d; font-size: 12px;">%s</p>
		<p style="color: #7f8c8d; font-size: 12px; margin-top: 30px;">This link will expire in 24 hours.</p>
		<p style="color: #7f8c8d; font-size: 12px;">If you didn't create an account, you can safely ignore this email.</p>
	</div>
</body>
</html>`, verificationURL, verificationURL)
}

// GenerateVerificationEmailText generates the plain text content for email verification
func GenerateVerificationEmailText(verificationURL string) string {
	return fmt.Sprintf(`Verify Your Email

Thank you for registering with Anki Backend!

Please click the link below to verify your email address:

%s

This link will expire in 24 hours.

If you didn't create an account, you can safely ignore this email.`, verificationURL)
}

