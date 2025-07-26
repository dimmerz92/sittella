package memorystore_test

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/dimmerz92/sittella/sessions"
	"github.com/dimmerz92/sittella/sessions/memorystore"
)

func newRequest(cookie *http.Cookie) (*httptest.ResponseRecorder, *http.Request) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)
	if cookie != nil {
		r.AddCookie(cookie)
	}
	return w, r
}

type testType struct {
	Data int
}

var key = "test-key"

var expected = testType{42}

func TestMemoryStore(t *testing.T) {
	store := memorystore.New(100*time.Millisecond, 90*time.Millisecond, sessions.CookieOpts{Name: "test"})
	defer store.Stop()

	w, r := newRequest(nil)

	t.Run("set", func(t *testing.T) {
		session := store.Session(w, r)

		if err := session.Set(key, expected); err != nil {
			t.Fatalf("failed to set data: %v", err)
		}

		if err := session.Save(); err != nil {
			t.Fatalf("failed to save session: %v", err)
		}

		if len(w.Result().Cookies()) != 1 && w.Result().Cookies()[0].Name != "test" {
			t.Fatal("failed to set cookie")
		}
	})

	t.Run("get", func(t *testing.T) {
		w, r = newRequest(w.Result().Cookies()[0])
		session := store.Session(w, r)

		var out testType
		if err := session.Get(key, &out); err != nil {
			t.Fatalf("failed to get data: %v", err)
		}

		if !reflect.DeepEqual(expected, out) {
			t.Fatalf("expected %#v got %#v", expected, out)
		}
	})

	t.Run("delete", func(t *testing.T) {
		w, r = newRequest(r.Cookies()[0])
		session := store.Session(w, r)

		session.Delete(key)

		if err := session.Save(); err != nil {
			t.Fatalf("failed to save session: %v", err)
		}

		var out testType
		if err := session.Get(key, &out); err != sessions.ErrValueNotFound {
			t.Fatalf("expected not found, got %v", err)
		}
	})

	t.Run("extended & expired", func(t *testing.T) {
		w, r = newRequest(r.Cookies()[0])
		session := store.Session(w, r)

		if err := session.Set(key, expected); err != nil {
			t.Fatalf("failed to set data: %v", err)
		}

		session.Extend(180 * time.Millisecond)

		if err := session.Save(); err != nil {
			t.Fatalf("failed to save session: %v", err)
		}

		time.Sleep(110 * time.Millisecond)

		w, r = newRequest(r.Cookies()[0])
		session = store.Session(w, r)

		var out testType
		if err := session.Get(key, &out); err != nil {
			t.Fatalf("failed to get data: %v", err)
		}

		time.Sleep(90 * time.Millisecond)

		w, r = newRequest(r.Cookies()[0])
		session = store.Session(w, r)

		var out2 testType
		if err := session.Get(key, &out2); err != sessions.ErrValueNotFound {
			t.Fatalf("expected not found, got %v", err)
		}
	})
}
