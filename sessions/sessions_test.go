package sessions_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/dimmerz92/sittella/sessions"
)

func newRequest(cookie *http.Cookie) (*httptest.ResponseRecorder, *http.Request) {
	r := httptest.NewRequest("GET", "/", nil)
	if cookie != nil {
		r.AddCookie(cookie)
	}
	return httptest.NewRecorder(), r
}

func TestMemorySessionStore(t *testing.T) {
	var afterSet, afterGet, afterDelete string
	config := sessions.MemorySessionStoreConfig{
		AfterSet:    func() { afterSet = "I was set" },
		AfterGet:    func() { afterGet = "I was retrieved" },
		AfterDelete: func() { afterDelete = "I was deleted" },
		Interval:    50 * time.Millisecond,
		TTL:         100 * time.Millisecond,
	}
	var sessionstore sessions.SessionStore = sessions.NewMemorySessionStore(config)
	defer sessionstore.Stop()

	key := "testkey"
	expected := "hello session!"

	t.Run("set", func(t *testing.T) {
		w, r := newRequest(nil)
		store := sessionstore.WithContext(w, r)

		if err := store.Set(key, expected); err != nil {
			t.Fatalf("expected to set session, got %v", err)
		}

		if afterSet != "I was set" {
			t.Fatalf("expected variable to be set, got %s", afterSet)
		}
	})

	t.Run("get", func(t *testing.T) {
		w, r := newRequest(nil)
		store := sessionstore.WithContext(w, r)

		var out string
		if err := store.Get(key, &out); err != nil {
			t.Fatalf("expected successful get, got %v", err)
		}

		if out != expected {
			t.Fatalf("expected %s, got %s", expected, out)
		}

		if afterGet != "I was retrieved" {
			t.Fatalf("expected variable to be set, got %s", afterGet)
		}
	})

	t.Run("delete", func(t *testing.T) {
		w, r := newRequest(nil)
		store := sessionstore.WithContext(w, r)

		if err := store.Delete(key); err != nil {
			t.Fatalf("expected delete, got %v", err)
		}

		if afterDelete != "I was deleted" {
			t.Fatalf("expected variable to be set, got %s", afterDelete)
		}

		var out string
		if err := store.Get(key, &out); err != sessions.ErrSessionNotFound {
			t.Fatalf("expected not found, got %v", err)
		}
	})

	t.Run("expire", func(t *testing.T) {
		w, r := newRequest(nil)
		store := sessionstore.WithContext(w, r)

		if err := store.Set(key, expected); err != nil {
			t.Fatalf("expected to set session, got %v", err)
		}

		time.Sleep(101 * time.Millisecond)

		w, r = newRequest(nil)
		store = sessionstore.WithContext(w, r)

		var out string
		if err := store.Get(key, &out); err != sessions.ErrSessionNotFound {
			t.Fatalf("expected not found, got %v", err)
		}
	})
}
