package mailer

import (
	"os"
	"time"

	"github.com/resend/resend-go/v2"
	"go.uber.org/zap"
	log "skhaz.dev/urlshortnen/logging"
)

type Mail struct {
	client *resend.Client
}

func NewMail() *Mail {
	return &Mail{
		client: resend.NewClient(os.Getenv("RESEND_APIKEY")),
	}
}

const (
	maxRetries = 3
	retryDelay = 3 * time.Second
)

func (m *Mail) Send(to, subject, html string) {
	params := &resend.SendEmailRequest{
		From:    "noreply@takealook.pro",
		To:      []string{to},
		Subject: subject,
		Html:    html,
	}

	for attempt := 1; attempt <= maxRetries; attempt++ {
		if _, err := m.client.Emails.Send(params); err != nil {
			log.Error("error while sending mail",
				zap.Int("attempt", attempt),
				zap.String("to", to),
				zap.Error(err),
			)

			if attempt == maxRetries {
				log.Warn("failed to send email after 3 attempts", zap.String("to", to))
			} else {
				time.Sleep(retryDelay)
			}
		} else {
			log.Info("email sent successfully", zap.String("to", to))
			break
		}
	}
}
