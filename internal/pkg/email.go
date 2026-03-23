package pkg

import (
	"bytes"
	"crypto/tls"
	"embed"
	"fmt"
	"html/template"
	"net"
	"net/smtp"
)

//go:embed templates/*
var emailTemplates embed.FS

type EmailService struct {
	host     string
	port     string
	username string
	password string
	from     string
}

func NewEmailService(host, port, username, password, from string) *EmailService {
	return &EmailService{
		host:     host,
		port:     port,
		username: username,
		password: password,
		from:     from,
	}
}

func (s *EmailService) SendPasswordResetEmail(to, resetLink string) error {
	body, err := s.renderTemplate("templates/reset_password.html", map[string]string{
		"ResetLink": resetLink,
	})
	if err != nil {
		return fmt.Errorf("render template: %w", err)
	}

	subject := "Recuperar contraseña - tepidolacuenta"
	msg := fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n%s",
		s.from, to, subject, body,
	)

	addr := fmt.Sprintf("%s:%s", s.host, s.port)

	if s.port == "465" {
		return s.sendWithTLS(addr, to, []byte(msg))
	}

	auth := smtp.PlainAuth("", s.username, s.password, s.host)
	return smtp.SendMail(addr, auth, s.from, []string{to}, []byte(msg))
}

func (s *EmailService) SendWelcomeEmail(to, loginLink string) error {
	body, err := s.renderTemplate("templates/welcome.html", map[string]string{
		"LoginLink": loginLink,
	})
	if err != nil {
		return fmt.Errorf("render template: %w", err)
	}

	subject := "Bienvenido/a a tepidolacuenta"
	msg := fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n%s",
		s.from, to, subject, body,
	)

	addr := fmt.Sprintf("%s:%s", s.host, s.port)

	if s.port == "465" {
		return s.sendWithTLS(addr, to, []byte(msg))
	}

	auth := smtp.PlainAuth("", s.username, s.password, s.host)
	return smtp.SendMail(addr, auth, s.from, []string{to}, []byte(msg))
}

func (s *EmailService) renderTemplate(path string, data any) (string, error) {
	tmpl, err := template.ParseFS(emailTemplates, path)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func (s *EmailService) sendWithTLS(addr, to string, msg []byte) error {
	tlsConfig := &tls.Config{
		ServerName: s.host,
	}

	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		return fmt.Errorf("tls dial: %w", err)
	}

	host, _, _ := net.SplitHostPort(addr)
	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return fmt.Errorf("smtp client: %w", err)
	}
	defer client.Close()

	auth := smtp.PlainAuth("", s.username, s.password, s.host)
	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("smtp auth: %w", err)
	}

	if err := client.Mail(s.from); err != nil {
		return fmt.Errorf("smtp mail from: %w", err)
	}

	if err := client.Rcpt(to); err != nil {
		return fmt.Errorf("smtp rcpt: %w", err)
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("smtp data: %w", err)
	}

	if _, err := w.Write(msg); err != nil {
		return fmt.Errorf("smtp write: %w", err)
	}

	return w.Close()
}
