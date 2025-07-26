package memorystore

import (
	"bytes"
	"context"
	"encoding/gob"
	"net/http"
	"sync"
	"time"

	"github.com/dimmerz92/sittella/sessions"
	"github.com/google/uuid"
)

type session struct {
	id     string
	expiry time.Time
	data   map[string][]byte
}

// Store defines an in memory session store to retrieve or generate request scoped sessions.
type Store struct {
	cancel context.CancelFunc
	cookie sessions.CookieOpts
	data   sync.Map
	ttl    time.Duration
}

// New returns a new in memory session Store.
// Parameters:
// - interval specifies the frequency that expired sessions are cleared.
// - ttl specifies the time to live for individual sessions.
// - opts specify the cookie options to be used.
func New(interval, ttl time.Duration, opts sessions.CookieOpts) *Store {
	ctx, cancel := context.WithCancel(context.Background())

	store := &Store{
		cancel: cancel,
		cookie: opts,
		data:   sync.Map{},
		ttl:    ttl,
	}

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return

			case <-ticker.C:
				store.data.Range(func(key, value any) bool {
					if time.Now().After(value.(session).expiry) {
						store.data.Delete(key)
					}
					return true
				})
			}
		}
	}()

	return store
}

// Session retrieves or generates a new session for the current request.
func (s *Store) Session(w http.ResponseWriter, r *http.Request) sessions.Session {
	sess := &Session{
		req:   r,
		res:   w,
		store: s,
		mu:    &sync.RWMutex{},
	}

	cookie, _ := r.Cookie(s.cookie.Name)
	if cookie != nil {
		data, ok := s.data.Load(cookie.Value)
		if ok && time.Now().Before(data.(session).expiry) {
			sess.data = data.(session)
			return sess
		}
	}

	sess.data = session{
		id:     uuid.NewString(),
		expiry: time.Now().Add(s.ttl),
		data:   make(map[string][]byte),
	}

	http.SetCookie(w, s.cookie.ToCookie(sess.data.id, s.ttl))

	return sess
}

// Stop releases any goroutines and resources allocated by the Store.
func (s *Store) Stop() { s.cancel() }

// Session defines an in memory request scoped session.
type Session struct {
	req     *http.Request
	res     http.ResponseWriter
	store   *Store
	data    session
	mu      *sync.RWMutex
	ttl     time.Duration
	changed bool
}

// Set adds or updates the key value pair to the session.
func (s *Session) Set(key string, value any) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(value); err != nil {
		return err
	}

	s.data.data[key] = buf.Bytes()

	s.changed = true

	return nil
}

// Get retrieves and decodes the value mapped to by the given key if it exists.
func (s *Session) Get(key string, dest any) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, ok := s.data.data[key]
	if !ok {
		return sessions.ErrValueNotFound
	}

	return gob.NewDecoder(bytes.NewBuffer(data)).Decode(dest)
}

// Delete removes the value mapped to by the given key from the session.
func (s *Session) Delete(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.data.data, key)

	s.changed = true
}

// Clear removes all key value pairs from the session.
func (s *Session) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.data.data = make(map[string][]byte)

	s.changed = true
}

// Expiry returns the expiry datetime for the session.
func (s *Session) Expiry() time.Time {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.data.expiry
}

// Extend updates the session expiry.
func (s *Session) Extend(ttl time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if ttl > 0 {
		s.ttl = ttl
		s.changed = true
	}
}

// Save persists any changes made to the session in the store.
func (s *Session) Save() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.changed {
		if s.ttl > 0 {
			s.data.expiry = time.Now().Add(s.ttl)
			http.SetCookie(s.res, s.store.cookie.ToCookie(s.data.id, s.ttl))
		}

		s.store.data.Store(s.data.id, s.data)
	}

	return nil
}
