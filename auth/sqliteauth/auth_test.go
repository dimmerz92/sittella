package sqliteauth_test

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/dimmerz92/sittella/auth"
	"github.com/dimmerz92/sittella/auth/sqliteauth"
	"github.com/dimmerz92/sittella/database/sqlitedb"
	"github.com/google/uuid"
)

func newRequest(cookie *http.Cookie) (*httptest.ResponseRecorder, *http.Request) {
	r := httptest.NewRequest("GET", "/", nil)
	if cookie != nil {
		r.AddCookie(cookie)
	}
	return httptest.NewRecorder(), r
}

func TestSQLiteAuth(t *testing.T) {
	db, err := sqlitedb.NewSQLiteDatabase(sqlitedb.Config{DSN: ":memory:"}, func(_ *sql.DB) int { return 0 })
	if err != nil {
		t.Fatalf("failed to open database")
	}

	config := sqliteauth.SQLiteAuthConfig{
		Database: db.DB(),
		Interval: 50 * time.Millisecond,
		TTL:      100 * time.Millisecond,
		Sliding:  true,
	}
	var authstore auth.Auth = sqliteauth.NewSQLiteAuth(config)
	defer authstore.Stop()

	w, r := newRequest(nil)

	expected := "hello auth!"

	t.Run("set", func(t *testing.T) {
		store := authstore.WithContext(w, r)

		if err := store.Set(expected); err != nil {
			t.Fatalf("expected to set auth, got %v", err)
		}

		if len(w.Result().Cookies()) != 1 {
			t.Fatalf("expected a cookie, got none")
		}
		cookie := w.Result().Cookies()[0]
		if _, err := uuid.Parse(cookie.Value); err != nil {
			t.Fatalf("expected uuid, got %v", err)
		}
	})

	t.Run("get", func(t *testing.T) {
		w, r := newRequest(w.Result().Cookies()[0])
		store := authstore.WithContext(w, r)

		var out string
		if err := store.Get(&out); err != nil {
			t.Fatalf("expected successful get, got %v", err)
		}

		if out != expected {
			t.Fatalf("expected %s, got %s", expected, out)
		}
	})

	t.Run("delete", func(t *testing.T) {
		w, r := newRequest(w.Result().Cookies()[0])
		store := authstore.WithContext(w, r)

		if err := store.Delete(); err != nil {
			t.Fatalf("expected delete, got %v", err)
		}

		var out string
		if err := store.Get(&out); err != auth.ErrAuthNotFound {
			t.Fatalf("expected not found, got %v", err)
		}
	})

	t.Run("slide and expire", func(t *testing.T) {
		w, r := newRequest(nil)
		store := authstore.WithContext(w, r)

		if err := store.Set(expected); err != nil {
			t.Fatalf("expected to set auth, got %v", err)
		}

		if len(w.Result().Cookies()) != 1 {
			t.Fatalf("expected a cookie, got none")
		}

		for range 5 {
			time.Sleep(90 * time.Millisecond)

			w, r = newRequest(w.Result().Cookies()[0])
			store = authstore.WithContext(w, r)

			var out string
			if err := store.Get(&out); err != nil {
				t.Fatalf("expected successful get, got %v", err)
			}

			if out != expected {
				t.Fatalf("expected %s, got %s", expected, out)
			}
		}

		time.Sleep(200 * time.Millisecond)

		w, r = newRequest(w.Result().Cookies()[0])
		store = authstore.WithContext(w, r)

		var out string
		if err := store.Get(&out); err != auth.ErrAuthNotFound {
			t.Fatalf("expected not found, got %v", err)
		}
	})
}
