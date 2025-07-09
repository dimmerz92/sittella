package sessions

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/dimmerz92/sittella/codecs"
)

type session struct {
	Expiry time.Time
	Data   []byte
}

// MemorySessionStoreContext provides request scoped access to the session store.
type MemorySessionStoreContext struct {
	res      http.ResponseWriter
	req      *http.Request
	sessions *MemorySessionStore
}

// Set adds the given data to the session store.
// May return backend specific errors or if the data is not serialisable by the configured codec.
func (s *MemorySessionStoreContext) Set(key string, data any, ttlOverride ...time.Duration) error {
	ttl := s.sessions.ttl
	if len(ttlOverride) > 0 && ttlOverride[0] > 0 {
		ttl = ttlOverride[0]
	}

	encoded, err := s.sessions.codec.Marshal(data)
	if err != nil {
		return err
	}

	s.sessions.store.Store(key, session{
		Expiry: time.Now().Add(ttl),
		Data:   encoded,
	})

	if s.sessions.afterSet != nil {
		s.sessions.afterSet()
	}

	return nil
}

// Get retrieves the data from the session store for the given key if it exists.
// Returns ErrSessionNotFound if n session is not found or is expired.
// May return backend specific errors or if the data is not serialisable by the configured codec.
func (s *MemorySessionStoreContext) Get(key string, dest any, ttlOverride ...time.Duration) error {
	data, ok := s.sessions.store.Load(key)
	if !ok {
		return ErrSessionNotFound
	}

	sess := data.(session)
	if time.Now().After(sess.Expiry) {
		s.sessions.store.Delete(key)
		return ErrSessionNotFound
	}

	if err := s.sessions.codec.Unmarshal(sess.Data, dest); err != nil {
		return err
	}

	if s.sessions.afterGet != nil {
		s.sessions.afterGet()
	}

	return nil
}

// Delete removes the data from the session store for the given key if it exists.
// May return backend specific errors.
func (s *MemorySessionStoreContext) Delete(key string) error {
	s.sessions.store.Delete(key)

	if s.sessions.afterDelete != nil {
		s.sessions.afterDelete()
	}

	return nil
}

// MemorySessionStore provides the router scoped in memory store to be distributed to contexts on each request.
type MemorySessionStore struct {
	afterSet    func()
	afterGet    func()
	afterDelete func()
	cancel      context.CancelFunc
	codec       codecs.Codec
	store       sync.Map
	ttl         time.Duration
}

// MemorySessionStoreConfig is used to configure an instance of the MemorySessionStore.
type MemorySessionStoreConfig struct {
	AfterSet    func()
	AfterGet    func()
	AfterDelete func()
	Codec       codecs.Codec
	Interval    time.Duration
	TTL         time.Duration
}

// NewMemorySessionStore instantiates an instance of the MemorySessionStore.
func NewMemorySessionStore(config MemorySessionStoreConfig) *MemorySessionStore {
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
	sessions := &MemorySessionStore{
		afterSet:    config.AfterSet,
		afterGet:    config.AfterGet,
		afterDelete: config.AfterDelete,
		cancel:      cancel,
		codec:       config.Codec,
		store:       sync.Map{},
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
				sessions.store.Range(func(key, value any) bool {
					if time.Now().After(value.(session).Expiry) {
						sessions.store.Delete(key)
					}
					return true
				})
			}
		}
	}()

	return sessions
}

// Returns a request scoped session store.
func (s *MemorySessionStore) WithContext(w http.ResponseWriter, r *http.Request) SessionStoreContext {
	return &MemorySessionStoreContext{
		res:      w,
		req:      r,
		sessions: s,
	}
}

// Stops any resources utilised by the session store.
// Note: this does not remove any data.
func (s *MemorySessionStore) Stop() error {
	s.cancel()
	return nil
}
