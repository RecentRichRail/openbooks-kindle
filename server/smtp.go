package server

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/smtp"
	"strings"
)

// SMTPService handles email sending functionality
type SMTPService struct {
	config *Config
}

// NewSMTPService creates a new SMTP service
func NewSMTPService(config *Config) *SMTPService {
	return &SMTPService{
		config: config,
	}
}

// SendBookToKindle sends a book file to a Kindle email address
func (s *SMTPService) SendBookToKindle(toEmail, bookTitle, author string, bookData io.Reader, filename string) error {
	if !s.config.SMTPEnabled {
		return fmt.Errorf("SMTP is not enabled")
	}

	// Create the email message
	subject := fmt.Sprintf("Book: %s by %s", bookTitle, author)
	
	// Read book data into memory
	bookBytes, err := io.ReadAll(bookData)
	if err != nil {
		return fmt.Errorf("failed to read book data: %w", err)
	}

	// Create MIME message with attachment
	message := s.createMIMEMessage(toEmail, subject, bookTitle, author, bookBytes, filename)

	// Send the email
	return s.sendEmail(toEmail, subject, message)
}

// sendEmail sends an email using SMTP
func (s *SMTPService) sendEmail(to, subject, message string) error {
	auth := smtp.PlainAuth("", s.config.SMTPUsername, s.config.SMTPPassword, s.config.SMTPHost)
	
	addr := fmt.Sprintf("%s:%d", s.config.SMTPHost, s.config.SMTPPort)
	
	return smtp.SendMail(addr, auth, s.config.SMTPFrom, []string{to}, []byte(message))
}

// createMIMEMessage creates a MIME email message with attachment
func (s *SMTPService) createMIMEMessage(to, subject, bookTitle, author string, bookData []byte, filename string) string {
	boundary := "openbooks-boundary-12345"
	
	var message strings.Builder
	
	// Email headers
	message.WriteString(fmt.Sprintf("To: %s\r\n", to))
	message.WriteString(fmt.Sprintf("From: %s\r\n", s.config.SMTPFrom))
	message.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	message.WriteString("MIME-Version: 1.0\r\n")
	message.WriteString(fmt.Sprintf("Content-Type: multipart/mixed; boundary=%s\r\n", boundary))
	message.WriteString("\r\n")
	
	// Email body
	message.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	message.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
	message.WriteString("\r\n")
	message.WriteString(fmt.Sprintf("Please find attached: %s by %s\r\n", bookTitle, author))
	message.WriteString("\r\nSent from OpenBooks\r\n")
	message.WriteString("\r\n")
	
	// Attachment
	message.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	message.WriteString(fmt.Sprintf("Content-Type: application/octet-stream\r\n"))
	message.WriteString(fmt.Sprintf("Content-Disposition: attachment; filename=\"%s\"\r\n", filename))
	message.WriteString("Content-Transfer-Encoding: base64\r\n")
	message.WriteString("\r\n")
	
	// Encode attachment in base64
	encoded := base64.StdEncoding.EncodeToString(bookData)
	// Add line breaks every 76 characters
	for i := 0; i < len(encoded); i += 76 {
		end := i + 76
		if end > len(encoded) {
			end = len(encoded)
		}
		message.WriteString(encoded[i:end])
		message.WriteString("\r\n")
	}
	
	// End boundary
	message.WriteString(fmt.Sprintf("--%s--\r\n", boundary))
	
	return message.String()
}
