package sqlitesessions

import (
	"context"
	"database/sql"
	"log/slog"
	"net/http"
	"time"

	"github.com/dimmerz92/sittella/codecs"
	"github.com/dimmerz92/sittella/sessions"
)

// SQLiteSessionStoreContext provides request scoped access to the session store.
type SQLiteSessionStoreContext struct {
	res      http.ResponseWriter
	req      *http.Request
	sessions *SQLiteSessionStore
}

const upsertSession = `INSERT INTO __sittella_sessions (id, expires_at, data) VALUES (?, ?, ?)
	ON CONFLICT(id) DO UPDATE SET expires_at = excluded.expires_at, data = excluded.data`

// Set adds the given data to the session store.
// May return backend specific errors or if the data is not serialisable by the configured codec.
func (s *SQLiteSessionStoreContext) Set(key string, data any, ttlOverride ...time.Duration) error {
	ttl := s.sessions.ttl
	if len(ttlOverride) > 0 && ttlOverride[0] > 0 {
		ttl = ttlOverride[0]
	}

	encoded, err := s.sessions.codec.Marshal(data)
	if err != nil {
		return err
	}

	if _, err := s.sessions.db.ExecContext(s.req.Context(), upsertSession, key, time.Now().Add(ttl), encoded); err != nil {
		return err
	}

	if s.sessions.afterSet != nil {
		s.sessions.afterSet()
	}

	return nil
}

const selectSession = "SELECT expires_at, data FROM __sittella_sessions WHERE id = ?"

// Get retrieves the data from the session store for the given key if it exists.
// Returns ErrSessionNotFound if n session is not found or is expired.
// May return backend specific errors or if the data is not serialisable by the configured codec.
func (s *SQLiteSessionStoreContext) Get(key string, dest any, ttlOverride ...time.Duration) error {
	var expiry time.Time
	var data []byte
	if err := s.sessions.db.QueryRowContext(s.req.Context(), selectSession, key).Scan(&expiry, &data); err != nil {
		return sessions.ErrSessionNotFound
	}

	if time.Now().After(expiry) {
		s.Delete(key)
		return sessions.ErrSessionNotFound
	}

	if err := s.sessions.codec.Unmarshal(data, dest); err != nil {
		return err
	}

	if s.sessions.afterGet != nil {
		s.sessions.afterGet()
	}

	return nil
}

const deleteSession = "DELETE FROM __sittella_sessions WHERE id = ?"

// Delete removes the data from the session store for the given key if it exists.
// May return backend specific errors.
func (s *SQLiteSessionStoreContext) Delete(key string) error {
	if _, err := s.sessions.db.ExecContext(s.req.Context(), deleteSession, key); err != nil {
		return err
	}

	if s.sessions.afterDelete != nil {
		s.sessions.afterDelete()
	}

	return nil
}

// SQLiteSessionStore provides the router scoped in sqlite store to be distributed to contexts on each request.
type SQLiteSessionStore struct {
	afterSet    func()
	afterGet    func()
	afterDelete func()
	cancel      context.CancelFunc
	codec       codecs.Codec
	db          *sql.DB
	ttl         time.Duration
}

// SQLiteSessionStoreConfig is used to configure an instance of the SQLiteSessionStore.
type SQLiteSessionStoreConfig struct {
	AfterSet    func()
	AfterGet    func()
	AfterDelete func()
	Database    *sql.DB
	Codec       codecs.Codec
	Interval    time.Duration
	TTL         time.Duration
}

const deleteExpiredSessions = "DELETE FROM __sittella_sessions WHERE expires_at <= ?"

// NewSQLiteSessionStore instantiates an instance of the SQLiteSessionStore.
func NewSQLiteSessionStore(config SQLiteSessionStoreConfig) *SQLiteSessionStore {
	if config.Database == nil {
		panic("database cannot be nil")
	}

	if config.Interval <= 0 {
		config.Interval = 10 * time.Minute
	}

	if config.TTL <= 0 {
		config.TTL = time.Hour
	}

	if config.Codec == nil {
		config.Codec = codecs.JSONCodec{}
	}

	ctx, cancel := context.WithCancel(context.Background())
	sessions := &SQLiteSessionStore{
		afterSet:    config.AfterSet,
		afterGet:    config.AfterGet,
		afterDelete: config.AfterDelete,
		cancel:      cancel,
		db:          config.Database,
		codec:       config.Codec,
		ttl:         config.TTL,
	}

	go func() {
		ticker := time.NewTicker(config.Interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return

			case <-ticker.C:
				if _, err := sessions.db.ExecContext(context.Background(), deleteExpiredSessions, time.Now()); err != nil {
					slog.Error(err.Error())
				}
			}
		}
	}()

	return sessions
}

// Returns a request scoped session store.
func (s *SQLiteSessionStore) WithContext(w http.ResponseWriter, r *http.Request) sessions.SessionStoreContext {
	return &SQLiteSessionStoreContext{
		res:      w,
		req:      r,
		sessions: s,
	}
}

// Stops any resources utilised by the session store.
// Note: this does not remove any data.
func (s *SQLiteSessionStore) Stop() error {
	s.cancel()
	return nil
}
