package mailer

import (
	"context"
	"log/slog"

	"github.com/dimmerz92/sittella/utils"
	"github.com/wneessen/go-mail"
)

// DefaultMailer is the default implementation of the Mailer interface.
// It wraps SMTP configuration and provides simple methods for sending email.
type DefaultMailer struct {
	// Host specifies the SMTP host (e.g. "smtp.example.com").
	// Use the Options to specify ports, TLS, and other options.
	Host string

	// From specifies the default "From" address used if Email.From is not set.
	From string

	// Options passed to the SMTP client (e.g. port, auth, TLS).
	Opts []mail.Option
}

// BuildEmail composes the given email into a send ready mail.Msg.
func (m *DefaultMailer) BuildEmail(email Email) (*mail.Msg, error) {
	message := mail.NewMsg()
	if err := message.From(utils.Coalesce(email.From, m.From)); err != nil {
		return nil, err
	}
	if email.ReplyTo != "" {
		if err := message.ReplyTo(email.ReplyTo); err != nil {
			return nil, err
		}
	}
	if err := message.To(email.ToRecipients()...); err != nil {
		return nil, err
	}
	if len(email.Cc) > 0 {
		if err := message.Cc(email.CcRecipients()...); err != nil {
			return nil, err
		}
	}
	if len(email.Bcc) > 0 {
		if err := message.Bcc(email.BccRecipients()...); err != nil {
			return nil, err
		}
	}
	for _, attachment := range email.Attachments {
		message.AttachReadSeeker(attachment.Filename, attachment.Content, attachment.Opts...)
	}
	message.SetMessageID()
	message.SetDate()
	message.Subject(email.Subject)
	message.SetBodyString(utils.Coalesce(email.BodyContentType, mail.TypeTextPlain), email.Body)
	return message, nil
}

// Send builds and sends a single email using the configured SMTP server.
func (m *DefaultMailer) Send(ctx context.Context, email Email) error {
	message, err := m.BuildEmail(email)
	if err != nil {
		return err
	}

	client, err := mail.NewClient(m.Host, m.Opts...)
	if err != nil {
		return err
	}

	return client.DialAndSendWithContext(ctx, message)
}

// SendMany sends multiple emails in bulk.
// Invalid messages are logged and skipped.
// Valid messages are sent together.
func (m *DefaultMailer) SendMany(ctx context.Context, emails []Email) error {
	var messages []*mail.Msg
	for _, email := range emails {
		message, err := m.BuildEmail(email)
		if err != nil {
			slog.Error(err.Error())
			continue
		}
		message.SetBulk()
		messages = append(messages, message)
	}

	client, err := mail.NewClient(m.Host, m.Opts...)
	if err != nil {
		return err
	}

	return client.DialAndSendWithContext(ctx, messages...)
}
