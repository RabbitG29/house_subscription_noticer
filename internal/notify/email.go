package notify

import (
	"fmt"
	"mime"
	"net/smtp"
	"strings"
)

// Email sends notifications through a plain SMTP account (Naver, Gmail, or
// any other provider that supports STARTTLS on submission). No external
// dependency is needed — net/smtp negotiates STARTTLS on its own when the
// server advertises it.
type Email struct {
	Host     string // e.g. "smtp.naver.com"
	Port     string // e.g. "587"
	Username string // full login, usually the sender address
	Password string // app password, not the account login password
	From     string
	To       string // comma-separated recipient list
}

func (e *Email) Send(subject, body string) error {
	recipients := splitAndTrim(e.To)
	if len(recipients) == 0 {
		return fmt.Errorf("수신자(EMAIL_TO)가 비어 있습니다")
	}

	msg := fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=\"UTF-8\"\r\n\r\n%s",
		e.From, strings.Join(recipients, ", "), mime.QEncoding.Encode("UTF-8", subject), body,
	)

	auth := smtp.PlainAuth("", e.Username, e.Password, e.Host)
	addr := e.Host + ":" + e.Port

	if err := smtp.SendMail(addr, auth, e.From, recipients, []byte(msg)); err != nil {
		return fmt.Errorf("smtp 발송 실패: %w", err)
	}
	return nil
}

func splitAndTrim(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
