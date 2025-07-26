package sessions

import (
	"errors"
	"net/http"
	"time"
)

var ErrValueNotFound = errors.New("value not found")

// Store specifies an interface to retrieve or generate request scoped sessions.
type Store interface {
	// Session retrieves or generates a new session for the current request.
	Session(w http.ResponseWriter, r *http.Request) Session

	// Stop releases any goroutines and resources allocated by the Store.
	Stop()
}

// Session specifies an interface for a request scoped session store.
type Session interface {
	// Set adds or updates the key value pair to the session.
	Set(key string, value any) error

	// Get retrieves and decodes the value mapped to by the given key if it exists.
	Get(key string, dest any) error

	// Delete removes the value mapped to by the given key from the session.
	Delete(key string)

	// Clear removes all key value pairs from the session.
	Clear()

	// Expiry returns the expiry datetime for the session.
	Expiry() time.Time

	// Extend updates the session expiry.
	Extend(ttl time.Duration)

	// Save persists any changes made to the session in the store.
	Save() error
}
