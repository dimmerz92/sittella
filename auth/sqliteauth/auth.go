package sqliteauth

import (
	"context"
	"database/sql"
	"log/slog"
	"net/http"
	"time"

	"github.com/dimmerz92/sittella/auth"
	"github.com/dimmerz92/sittella/codecs"
)

// SQLiteAuthContext provides request scoped access to the auth store.
type SQLiteAuthContext struct {
	res  http.ResponseWriter
	req  *http.Request
	auth *SQLiteAuth
}

const insertAuth = "INSERT INTO __sittella_auth (expires_at, data) VALUES (?, ?) RETURNING id"

// Set adds the given data to the auth store and sets a session cookie.
// May return backend specific errors or if the data is not serialisable by the configured codec.
func (a *SQLiteAuthContext) Set(data any, ttlOverride ...time.Duration) error {
	ttl := a.auth.ttl
	if len(ttlOverride) > 0 && ttlOverride[0] > 0 {
		ttl = ttlOverride[0]
	}

	encoded, err := a.auth.codec.Marshal(data)
	if err != nil {
		return err
	}

	var sessionId string
	if err := a.auth.db.QueryRowContext(a.req.Context(), insertAuth, time.Now().Add(ttl), encoded).Scan(&sessionId); err != nil {
		return err
	}
	auth.SetSessionCookie(a.res, sessionId, ttl, a.auth.cookieOpts)

	return nil
}

const selectAuth = "SELECT expires_at, data FROM __sittella_auth WHERE id = ?"
const updateAuth = "UPDATE __sittella_auth SET expires_at = ?"

// Get retrieves the data from the auth store if a session cookie exists.
// Returns ErrAuthNotFound if an auth session is not found or is expired.
// May return backend specific errors or if the data is not serialisable by the configured codec.
func (a *SQLiteAuthContext) Get(dest any, ttlOverride ...time.Duration) error {
	sessionId, ok := auth.GetSessionIdFromCookie(a.req, a.auth.cookieOpts)
	if !ok {
		return auth.ErrAuthNotFound
	}

	var expiry time.Time
	var data []byte
	if err := a.auth.db.QueryRowContext(a.req.Context(), selectAuth, sessionId).Scan(&expiry, &data); err != nil {
		return auth.ErrAuthNotFound
	}

	if time.Now().After(expiry) {
		a.Delete()
		return auth.ErrAuthNotFound
	}

	if err := a.auth.codec.Unmarshal(data, dest); err != nil {
		return err
	}

	if a.auth.sliding {
		ttl := a.auth.ttl
		if len(ttlOverride) > 0 && ttlOverride[0] > 0 {
			ttl = ttlOverride[0]
		}
		if _, err := a.auth.db.ExecContext(a.req.Context(), updateAuth, time.Now().Add(ttl)); err != nil {
			return err
		}
		auth.SetSessionCookie(a.res, sessionId, ttl, a.auth.cookieOpts)
	}

	return nil
}

const deleteAuth = "DELETE FROM __sittella_auth WHERE id = ?"

// Delete removes the data from the auth store and revokes the session cookie if it exists.
// May return backend specific errors.
func (a *SQLiteAuthContext) Delete() error {
	var err error
	if sessionId, ok := auth.GetSessionIdFromCookie(a.req, a.auth.cookieOpts); ok {
		_, err = a.auth.db.ExecContext(a.req.Context(), deleteAuth, sessionId)
		auth.RevokeSessionCookie(a.res, a.auth.cookieOpts)
	}
	return err
}

// SQLiteAuth provides the router scoped sqlite auth store to be distributed to contexts on each request.
type SQLiteAuth struct {
	cancel     context.CancelFunc
	codec      codecs.Codec
	cookieOpts auth.CookieOpts
	sliding    bool
	db         *sql.DB
	ttl        time.Duration
}

// SQLiteAuthConfig is used to configure an instance of the SQLiteAuth.
type SQLiteAuthConfig struct {
	Codec      codecs.Codec
	CookieOpts auth.CookieOpts
	Database   *sql.DB
	Sliding    bool
	Interval   time.Duration
	TTL        time.Duration
}

const deleteExpiredAuth = "DELETE FROM __sittella_auth WHERE expires_at <= ?"

// NewSQLiteAuth instantiates an instance of the SQLiteSessionStore.
func NewSQLiteAuth(config SQLiteAuthConfig) *SQLiteAuth {
	if config.Database == nil {
		panic("database cannot be nil")
	}

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
	auth := &SQLiteAuth{
		ttl:        config.TTL,
		cancel:     cancel,
		codec:      config.Codec,
		cookieOpts: config.CookieOpts,
		sliding:    config.Sliding,
		db:         config.Database,
	}

	go func() {
		ticker := time.NewTicker(config.Interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return

			case <-ticker.C:
				if _, err := auth.db.ExecContext(ctx, deleteExpiredAuth, time.Now()); err != nil {
					slog.Error(err.Error())
				}
			}
		}
	}()

	return auth
}

// Returns a request scoped auth.
func (a *SQLiteAuth) WithContext(w http.ResponseWriter, r *http.Request) auth.AuthContext {
	return &SQLiteAuthContext{
		res:  w,
		req:  r,
		auth: a,
	}
}

// Stops any resources utilised by the auth.
// Note: this does not remove any data.
func (a *SQLiteAuth) Stop() error {
	a.cancel()
	return nil
}
