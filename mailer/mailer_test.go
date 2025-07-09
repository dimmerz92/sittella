package mailer_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/dimmerz92/sittella/mailer"
	"github.com/wneessen/go-mail"
)

type MailHogMessage struct {
	Content struct {
		Headers map[string][]string `json:"Headers"`
		Body    string              `json:"Body"`
	} `json:"Content"`
}

type MailHogResponse struct {
	Total int              `json:"total"`
	Items []MailHogMessage `json:"items"`
}

func clearMailHog(t *testing.T) {
	req, err := http.NewRequest("DELETE", "http://localhost:8025/api/v1/messages", nil)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("failed to clear MailHog messages: %s", resp.Status)
	}
}

func getLastMailHogMessages(t *testing.T, count int) []MailHogMessage {
	resp, err := http.Get("http://localhost:8025/api/v2/messages?limit=10")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	var data MailHogResponse
	if err := json.Unmarshal(body, &data); err != nil {
		t.Fatal(err)
	}

	if len(data.Items) < count {
		t.Fatalf("expected at least %d messages, got %d", count, len(data.Items))
	}

	return data.Items[:count]
}

func containsHeaderValue(headers map[string][]string, key, val string) bool {
	values, ok := headers[key]
	if !ok {
		return false
	}
	for _, v := range values {
		if strings.Contains(v, val) {
			return true
		}
	}
	return false
}

var email1 = mailer.Email{
	To:          []mailer.Recipient{{Email: "recipient@example.com"}},
	Subject:     "Test",
	Attachments: []mailer.Attachment{{Filename: "hello.txt", Content: bytes.NewReader([]byte("hello world!"))}},
	Body:        "Hello from test1!",
}

var email2 = mailer.Email{
	To:              []mailer.Recipient{{Email: "recipient@example.com"}},
	Subject:         "Test 2",
	Body:            `<style>p{font-size: 20px; color: blue;}</style><p>Hello from test2!</p>`,
	BodyContentType: mail.TypeTextHTML,
}

func TestMailer(t *testing.T) {
	emailer := &mailer.DefaultMailer{
		Host: "localhost",
		From: "test@example.com",
		Opts: []mail.Option{mail.WithPort(1025), mail.WithTLSPolicy(mail.NoTLS)},
	}

	t.Run("send", func(t *testing.T) {
		clearMailHog(t)

		if err := emailer.Send(t.Context(), email1); err != nil {
			t.Fatalf("Send() failed: %v", err)
		}

		time.Sleep(100 * time.Millisecond)

		msg := getLastMailHogMessages(t, 1)[0]
		if !containsHeaderValue(msg.Content.Headers, "From", "test@example.com") {
			t.Error("From header mismatch")
		}

		if !containsHeaderValue(msg.Content.Headers, "To", "recipient@example.com") {
			t.Error("To header mismatch")
		}

		if !containsHeaderValue(msg.Content.Headers, "Subject", "Test") {
			t.Error("Subject header mismatch")
		}
		if !strings.Contains(msg.Content.Body, "Hello from test") {
			t.Error("Email body content mismatch")
		}
	})

	t.Run("sendmany", func(t *testing.T) {
		clearMailHog(t)

		if err := emailer.SendMany(t.Context(), []mailer.Email{email1, email2}); err != nil {
			t.Fatalf("SendMany() failed: %v", err)
		}

		time.Sleep(100 * time.Millisecond)

		for _, msg := range getLastMailHogMessages(t, 2) {
			if !containsHeaderValue(msg.Content.Headers, "From", "test@example.com") {
				t.Error("From header mismatch")
			}

			if !containsHeaderValue(msg.Content.Headers, "To", "recipient@example.com") {
				t.Error("To header mismatch")
			}

			if !containsHeaderValue(msg.Content.Headers, "Subject", "Test") {
				t.Error("Subject header mismatch")
			}

			if !strings.Contains(msg.Content.Body, "Hello from test") {
				t.Error("Email body content mismatch")
			}
		}
	})
}
