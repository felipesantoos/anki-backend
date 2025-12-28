package email

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/smtp"
	"time"

	"github.com/felipesantos/anki-backend/config"
)

// SMTPRepository implements IEmailRepository using SMTP
type SMTPRepository struct {
	cfg config.EmailConfig
}

// NewSMTPRepository creates a new SMTP email repository
func NewSMTPRepository(cfg config.EmailConfig) (*SMTPRepository, error) {
	if !cfg.Enabled {
		return nil, fmt.Errorf("email is not enabled")
	}

	if cfg.SMTPHost == "" {
		return nil, fmt.Errorf("SMTP host is required")
	}

	if cfg.SMTPUser == "" {
		return nil, fmt.Errorf("SMTP user is required")
	}

	if cfg.SMTPPassword == "" {
		return nil, fmt.Errorf("SMTP password is required")
	}

	return &SMTPRepository{
		cfg: cfg,
	}, nil
}

// SendEmail sends an email via SMTP
func (r *SMTPRepository) SendEmail(ctx context.Context, to, subject, htmlBody, textBody string) error {
	// Set up authentication
	auth := smtp.PlainAuth("", r.cfg.SMTPUser, r.cfg.SMTPPassword, r.cfg.SMTPHost)

	// Build email message
	from := fmt.Sprintf("%s <%s>", r.cfg.FromName, r.cfg.FromEmail)
	headers := make(map[string]string)
	headers["From"] = from
	headers["To"] = to
	headers["Subject"] = subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "multipart/alternative; boundary=\"boundary123\""

	// Build message body with multipart/alternative
	message := ""
	for k, v := range headers {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n"
	message += "--boundary123\r\n"
	message += "Content-Type: text/plain; charset=UTF-8\r\n"
	message += "Content-Transfer-Encoding: quoted-printable\r\n"
	message += "\r\n"
	message += textBody
	message += "\r\n"
	message += "--boundary123\r\n"
	message += "Content-Type: text/html; charset=UTF-8\r\n"
	message += "Content-Transfer-Encoding: quoted-printable\r\n"
	message += "\r\n"
	message += htmlBody
	message += "\r\n"
	message += "--boundary123--\r\n"

	// Create SMTP address
	addr := fmt.Sprintf("%s:%d", r.cfg.SMTPHost, r.cfg.SMTPPort)

	// Create a context with timeout
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Send email in a goroutine to respect context cancellation
	errChan := make(chan error, 1)
	go func() {
		var err error
		if r.cfg.UseTLS {
			// Use TLS (STARTTLS)
			err = r.sendWithTLS(addr, auth, from, []string{to}, []byte(message))
		} else {
			// Use SSL (direct TLS connection)
			err = r.sendWithSSL(addr, auth, from, []string{to}, []byte(message))
		}
		errChan <- err
	}()

	select {
	case <-ctxWithTimeout.Done():
		return fmt.Errorf("email send timeout: %w", ctxWithTimeout.Err())
	case err := <-errChan:
		if err != nil {
			return fmt.Errorf("failed to send email: %w", err)
		}
		return nil
	}
}

// sendWithTLS sends email using STARTTLS (port 587)
func (r *SMTPRepository) sendWithTLS(addr string, auth smtp.Auth, from string, to []string, msg []byte) error {
	// Connect to SMTP server
	client, err := smtp.Dial(addr)
	if err != nil {
		return fmt.Errorf("failed to connect to SMTP server: %w", err)
	}
	defer client.Close()

	// Send STARTTLS command
	if ok, _ := client.Extension("STARTTLS"); ok {
		config := &tls.Config{ServerName: r.cfg.SMTPHost}
		if err := client.StartTLS(config); err != nil {
			return fmt.Errorf("failed to start TLS: %w", err)
		}
	}

	// Authenticate
	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("failed to authenticate: %w", err)
	}

	// Set sender and recipients
	if err := client.Mail(from); err != nil {
		return fmt.Errorf("failed to set sender: %w", err)
	}

	for _, recipient := range to {
		if err := client.Rcpt(recipient); err != nil {
			return fmt.Errorf("failed to set recipient %s: %w", recipient, err)
		}
	}

	// Send email body
	writer, err := client.Data()
	if err != nil {
		return fmt.Errorf("failed to open data writer: %w", err)
	}

	_, err = writer.Write(msg)
	if err != nil {
		writer.Close()
		return fmt.Errorf("failed to write email body: %w", err)
	}

	err = writer.Close()
	if err != nil {
		return fmt.Errorf("failed to close data writer: %w", err)
	}

	return client.Quit()
}

// sendWithSSL sends email using direct TLS connection (port 465)
func (r *SMTPRepository) sendWithSSL(addr string, auth smtp.Auth, from string, to []string, msg []byte) error {
	// Create TLS config
	tlsConfig := &tls.Config{
		ServerName: r.cfg.SMTPHost,
	}

	// Connect with TLS
	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		return fmt.Errorf("failed to connect to SMTP server with TLS: %w", err)
	}
	defer conn.Close()

	// Create SMTP client
	client, err := smtp.NewClient(conn, r.cfg.SMTPHost)
	if err != nil {
		return fmt.Errorf("failed to create SMTP client: %w", err)
	}
	defer client.Close()

	// Authenticate
	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("failed to authenticate: %w", err)
	}

	// Set sender and recipients
	if err := client.Mail(from); err != nil {
		return fmt.Errorf("failed to set sender: %w", err)
	}

	for _, recipient := range to {
		if err := client.Rcpt(recipient); err != nil {
			return fmt.Errorf("failed to set recipient %s: %w", recipient, err)
		}
	}

	// Send email body
	writer, err := client.Data()
	if err != nil {
		return fmt.Errorf("failed to open data writer: %w", err)
	}

	_, err = writer.Write(msg)
	if err != nil {
		writer.Close()
		return fmt.Errorf("failed to write email body: %w", err)
	}

	err = writer.Close()
	if err != nil {
		return fmt.Errorf("failed to close data writer: %w", err)
	}

	return client.Quit()
}

