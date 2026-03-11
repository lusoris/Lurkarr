package notifications

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"strings"
	"time"
)

// Email sends notifications via SMTP.
type Email struct {
	Host       string
	Port       int
	Username   string
	Password   string
	From       string
	To         []string
	StartTLS   bool
	SkipVerify bool
}

// NewEmail creates an Email notification provider.
func NewEmail(host string, port int, username, password, from string, to []string, startTLS, skipVerify bool) *Email {
	return &Email{
		Host:       host,
		Port:       port,
		Username:   username,
		Password:   password,
		From:       from,
		To:         to,
		StartTLS:   startTLS,
		SkipVerify: skipVerify,
	}
}

func (e *Email) Name() string { return "Email" }

func (e *Email) Send(ctx context.Context, event Event) error {
	subject := event.Title
	body := formatPlainMessage(event)

	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=utf-8\r\n\r\n%s",
		e.From,
		strings.Join(e.To, ","),
		subject,
		body,
	)

	addr := fmt.Sprintf("%s:%d", e.Host, e.Port)

	dialer := &net.Dialer{Timeout: 10 * time.Second}
	conn, err := dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		return fmt.Errorf("email dial failed: %w", err)
	}

	client, err := smtp.NewClient(conn, e.Host)
	if err != nil {
		conn.Close()
		return fmt.Errorf("email smtp client: %w", err)
	}
	defer client.Close()

	if e.StartTLS {
		tlsConfig := &tls.Config{
			ServerName: e.Host,
			MinVersion: tls.VersionTLS12,
		}
		if e.SkipVerify {
			tlsConfig.InsecureSkipVerify = true //nolint:gosec // user-configured
		}
		if err := client.StartTLS(tlsConfig); err != nil {
			return fmt.Errorf("email starttls: %w", err)
		}
	}

	if e.Username != "" {
		auth := smtp.PlainAuth("", e.Username, e.Password, e.Host)
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("email auth: %w", err)
		}
	}

	if err := client.Mail(e.From); err != nil {
		return fmt.Errorf("email MAIL FROM: %w", err)
	}
	for _, to := range e.To {
		if err := client.Rcpt(to); err != nil {
			return fmt.Errorf("email RCPT TO %s: %w", to, err)
		}
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("email DATA: %w", err)
	}
	if _, err := w.Write([]byte(msg)); err != nil {
		return fmt.Errorf("email write: %w", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("email close data: %w", err)
	}
	return client.Quit()
}

func (e *Email) Test(ctx context.Context) error {
	return e.Send(ctx, Event{
		Type:    EventTestNotification,
		Title:   "Lurkarr Test Notification",
		Message: "If you see this, email notifications are working!",
	})
}
