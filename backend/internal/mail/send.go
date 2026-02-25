package mail

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
	"strings"

	"github.com/xiaoxinpro/xxjz-go-web/backend/internal/config"
)

// Send sends an email using cfg. If cfg.Host is empty, returns error (mail not configured).
func Send(cfg *config.MailConfig, to, subject, body string) error {
	if cfg == nil || cfg.Host == "" {
		return fmt.Errorf("mail not configured")
	}
	addr := cfg.Host + ":" + cfg.Port
	if cfg.Port == "" {
		addr = cfg.Host + ":465"
	}
	msg := []byte("To: " + to + "\r\n" +
		"From: " + cfg.FromName + " <" + cfg.From + ">\r\n" +
		"Subject: " + subject + "\r\n" +
		"Content-Type: text/html; charset=UTF-8\r\n" +
		"\r\n" + body + "\r\n")

	if cfg.Secure == "ssl" || cfg.Secure == "tls" {
		tlsConfig := &tls.Config{ServerName: cfg.Host}
		conn, err := tls.Dial("tcp", addr, tlsConfig)
		if err != nil {
			return err
		}
		defer conn.Close()
		client, err := smtp.NewClient(conn, cfg.Host)
		if err != nil {
			return err
		}
		defer client.Close()
		if cfg.Username != "" && cfg.Password != "" {
			if err = client.Auth(smtp.PlainAuth("", cfg.Username, cfg.Password, cfg.Host)); err != nil {
				return err
			}
		}
		if err = client.Mail(cfg.From); err != nil {
			return err
		}
		if err = client.Rcpt(to); err != nil {
			return err
		}
		w, err := client.Data()
		if err != nil {
			return err
		}
		_, err = w.Write(msg)
		if err != nil {
			return err
		}
		return w.Close()
	}
	// no auth / plain
	auth := smtp.PlainAuth("", cfg.Username, cfg.Password, cfg.Host)
	return smtp.SendMail(addr, auth, cfg.From, []string{to}, msg)
}

// IsConfigured returns true if mail can be sent (host and from set).
func IsConfigured(cfg *config.MailConfig) bool {
	return cfg != nil && strings.TrimSpace(cfg.Host) != "" && strings.TrimSpace(cfg.From) != ""
}
