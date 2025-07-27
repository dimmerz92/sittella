package mailer

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/wneessen/go-mail"
)

// Mailer defines the interface for sending emails.
type Mailer interface {
	// Send sends a single email message.
	Send(ctx context.Context, email Email) error

	// SendMany sends multiple email messages in a batch.
	SendMany(ctx context.Context, emails []Email) error
}

// Attachment represents a single email attachment,
// including its filename, MIME content type, and raw byte content.
type Attachment struct {
	// Filename specifies the name of the attached file.
	Filename string

	// Content specifies the file to be attached.
	Content io.ReadSeeker

	// Opts specifies any optional file options to apply.
	Opts []mail.FileOption
}

type Recipient struct {
	// Name specifies the recipient's name.
	Name string

	// Email specifies the recipient's email.
	Email string
}

// Email represents the structure of an email message.
// It supports multiple recipients, optional reply-to, attachments, and custom headers.
type Email struct {
	// From specifies the email address of the sender.
	From string

	// ReplyTo specifies an optional reply to address; if empty, defaults to From.
	ReplyTo string

	// To specifies a list of primary recipient email addresses.
	To []Recipient

	// Cc specifies a list of carbon copy recipient email addresses.
	Cc []Recipient

	// Bcc specifies list of blind carbon copy recipient email addresses.
	Bcc []Recipient

	// Subject specifies the subject line of the email.
	Subject string

	// Attachments specifies an optional list of file attachments.
	Attachments []Attachment

	// Body specifies the email body content.
	Body string

	// BodyContentType specifies the body content type.
	// Defaults to plain text if not set.
	BodyContentType mail.ContentType
}

// Format returns a recipient as an email formatted string.
// E.g. `John Doe <john.doe@email.com>` or `john.doe@email.com`
func (e *Email) Format(recipient Recipient) string {
	name := strings.TrimSpace(recipient.Name)
	email := strings.TrimSpace(recipient.Email)

	if name != "" {
		return fmt.Sprintf("%s <%s>", name, email)
	}

	return email
}

// ToRecipients returns a list of formatted primary recipient strings.
func (e *Email) ToRecipients() []string {
	var recipients []string
	for _, recipient := range e.To {
		recipients = append(recipients, e.Format(recipient))
	}
	return recipients
}

// CcRecipients returns a list of formatted CC recipient strings.
func (e *Email) CcRecipients() []string {
	var recipients []string
	for _, recipient := range e.Cc {
		recipients = append(recipients, e.Format(recipient))
	}
	return recipients
}

// BccRecipients returns a list of formatted BCC recipient strings.
func (e *Email) BccRecipients() []string {
	var recipients []string
	for _, recipient := range e.Bcc {
		recipients = append(recipients, e.Format(recipient))
	}
	return recipients
}
