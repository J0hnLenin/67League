package email

import (
	"context"
	"fmt"
	"log/slog"
	"net/smtp"

	"github.com/J0hnLenin/67League/internal/application/ports"
	"github.com/J0hnLenin/67League/internal/domain/shared"
)

// SMTPConfig — настройки SMTP.
type SMTPConfig struct {
	Host     string
	Port     string
	Username string
	Password string
	From     string
}

// SMTPSender реализует ports.EmailSender.
type SMTPSender struct {
	cfg SMTPConfig
}

func NewSMTPSender(cfg SMTPConfig) *SMTPSender { return &SMTPSender{cfg: cfg} }

func (s *SMTPSender) Send(ctx context.Context, to shared.Email, subject, body string) error {
	msg := []byte("From: " + s.cfg.From + "\r\n" +
		"To: " + string(to) + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/plain; charset=UTF-8\r\n\r\n" +
		body + "\r\n")

	auth := smtp.PlainAuth("", s.cfg.Username, s.cfg.Password, s.cfg.Host)
	addr := s.cfg.Host + ":" + s.cfg.Port
	if err := smtp.SendMail(addr, auth, s.cfg.From, []string{string(to)}, msg); err != nil {
		return fmt.Errorf("send email: %w", err)
	}
	return nil
}

func (s *SMTPSender) SendTemplate(ctx context.Context, to shared.Email, template string, data map[string]any) error {
	subject, body := renderTemplate(template, data)
	return s.Send(ctx, to, subject, body)
}

// LogSender — простой dev-отправитель, пишет письма в лог.
// Используется в тестах и локально, когда SMTP не настроен.
type LogSender struct {
	Logger *slog.Logger
}

func (s *LogSender) Send(ctx context.Context, to shared.Email, subject, body string) error {
	s.Logger.Info("email",
		"to", string(to),
		"subject", subject,
		"body", body,
	)
	return nil
}

func (s *LogSender) SendTemplate(ctx context.Context, to shared.Email, template string, data map[string]any) error {
	subject, body := renderTemplate(template, data)
	return s.Send(ctx, to, subject, body)
}

func renderTemplate(name string, data map[string]any) (subject, body string) {
	switch name {
	case "new_challenge":
		return "6-7 Лига: новый вызов", fmt.Sprintf("Вам отправили вызов на партию.\n%v\n", data)
	case "challenge_accepted":
		return "6-7 Лига: вызов принят", fmt.Sprintf("Соперник принял ваш вызов.\n%v\n", data)
	case "result_submitted":
		return "6-7 Лига: результат внесён", fmt.Sprintf("Внесите подтверждение результата.\n%v\n", data)
	case "result_confirmed":
		return "6-7 Лига: результат подтверждён", fmt.Sprintf("Партия подтверждена.\n%v\n", data)
	case "badge_awarded":
		return "6-7 Лига: новый бейдж", fmt.Sprintf("Вы получили бейдж.\n%v\n", data)
	default:
		return "6-7 Лига", fmt.Sprintf("%s\n%v\n", name, data)
	}
}

// Компилятор сам проверит, что мы реализуем интерфейс.
var _ ports.EmailSender = (*SMTPSender)(nil)
var _ ports.EmailSender = (*LogSender)(nil)
