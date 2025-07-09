package auth

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/dimmerz92/sittella/codecs"
	"github.com/google/uuid"
)

type session struct {
	Expiry time.Time
	Data   []byte
}

// MemoryAuthContext provides request scoped access to the auth store.
type MemoryAuthContext struct {
	res  http.ResponseWriter
	req  *http.Request
	auth *MemoryAuth
}

// Set adds the given data to the auth store and sets a session cookie.
// May return backend specific errors or if the data is not serialisable by the configured codec.
func (a *MemoryAuthContext) Set(data any, ttlOverride ...time.Duration) error {
	sessionId := uuid.NewString()
	ttl := a.auth.ttl
	if len(ttlOverride) > 0 && ttlOverride[0] > 0 {
		ttl = ttlOverride[0]
	}

	encoded, err := a.auth.codec.Marshal(data)
	if err != nil {
		return err
	}

	a.auth.store.Store(sessionId, session{
		Expiry: time.Now().Add(ttl),
		Data:   encoded,
	})
	SetSessionCookie(a.res, sessionId, ttl, a.auth.cookieOpts)

	return nil
}

// Get retrieves the data from the auth store if a session cookie exists.
// Returns ErrAuthNotFound if an auth session is not found or is expired.
// May return backend specific errors or if the data is not serialisable by the configured codec.
func (a *MemoryAuthContext) Get(dest any, ttlOverride ...time.Duration) error {
	sessionId, ok := GetSessionIdFromCookie(a.req, a.auth.cookieOpts)
	if !ok {
		return ErrAuthNotFound
	}

	data, ok := a.auth.store.Load(sessionId)
	if !ok {
		return ErrAuthNotFound
	}

	sess := data.(session)
	if time.Now().After(sess.Expiry) {
		a.auth.store.Delete(sessionId)
		return ErrAuthNotFound
	}

	if err := a.auth.codec.Unmarshal(sess.Data, dest); err != nil {
		return err
	}

	if a.auth.sliding {
		ttl := a.auth.ttl
		if len(ttlOverride) > 0 && ttlOverride[0] > 0 {
			ttl = ttlOverride[0]
		}
		a.auth.store.Store(sessionId, session{
			Expiry: time.Now().Add(ttl),
			Data:   sess.Data,
		})
		SetSessionCookie(a.res, sessionId, ttl, a.auth.cookieOpts)
	}

	return nil
}

// Delete removes the data from the auth store and revokes the session cookie if it exists.
// May return backend specific errors.
func (a *MemoryAuthContext) Delete() error {
	if sessionId, ok := GetSessionIdFromCookie(a.req, a.auth.cookieOpts); ok {
		a.auth.store.Delete(sessionId)
		RevokeSessionCookie(a.res, a.auth.cookieOpts)
	}
	return nil
}

// MemoryAuth provides the router scoped in memory auth store to be distributed to contexts on each request.
type MemoryAuth struct {
	cancel     context.CancelFunc
	codec      codecs.Codec
	cookieOpts CookieOpts
	sliding    bool
	store      sync.Map
	ttl        time.Duration
}

// MemoryAuthConfig is used to configure an instance of the MemoryAuth.
type MemoryAuthConfig struct {
	Codec      codecs.Codec
	CookieOpts CookieOpts
	Sliding    bool
	Interval   time.Duration
	TTL        time.Duration
}

// NewMemoryAuth instantiates an instance of the MemorySessionStore.
func NewMemoryAuth(config MemoryAuthConfig) *MemoryAuth {
	if config.Interval <= 0 {
		config.Interval = 10 * time.Minute
	}

	if config.TTL <= 0 {
		config.TTL = 24 * time.Hour
	}

	if config.Codec == nil {
		config.Codec = codecs.JSONCodec{}
	}

	ctx, cancel := context.WithCancel(context.Background())
	auth := &MemoryAuth{
		ttl:        config.TTL,
		cancel:     cancel,
		codec:      config.Codec,
		cookieOpts: config.CookieOpts,
		sliding:    config.Sliding,
		store:      sync.Map{},
	}

	go func() {
		ticker := time.NewTicker(config.Interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return

			case <-ticker.C:
				auth.store.Range(func(key, value any) bool {
					if time.Now().After(value.(session).Expiry) {
						auth.store.Delete(key)
					}
					return true
				})
			}
		}
	}()

	return auth
}

// Returns a request scoped auth.
func (a *MemoryAuth) WithContext(w http.ResponseWriter, r *http.Request) AuthContext {
	return &MemoryAuthContext{
		res:  w,
		req:  r,
		auth: a,
	}
}

// Stops any resources utilised by the auth.
// Note: this does not remove any data.
func (a *MemoryAuth) Stop() error {
	a.cancel()
	return nil
}
