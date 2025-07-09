package sessions

import (
	"errors"
	"net/http"
	"time"
)

var ErrSessionNotFound = errors.New("session not found")

// SessionStoreContext defines the sittella request scoped session store.
type SessionStoreContext interface {
	// Set adds the given data to the session store.
	// May return backend specific errors or if the data is not serialisable by the configured codec.
	Set(key string, data any, ttlOverride ...time.Duration) error

	// Get retrieves the data from the session store for the given key if it exists.
	// Returns ErrSessionNotFound if n session is not found or is expired.
	// May return backend specific errors or if the data is not serialisable by the configured codec.
	Get(key string, dest any, ttlOverride ...time.Duration) error

	// Delete removes the data from the session store for the given key if it exists.
	// May return backend specific errors.
	Delete(key string) error
}

// SessionStore defines the sittella application scoped session store.
type SessionStore interface {
	// Returns a request scoped session store.
	WithContext(w http.ResponseWriter, r *http.Request) SessionStoreContext

	// Stops any resources utilised by the session store.
	// Note: this does not remove any data.
	Stop() error
}
