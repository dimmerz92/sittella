package auth

import (
	"errors"
	"net/http"
	"time"
)

var ErrAuthNotFound = errors.New("auth not found")

// AuthContext defines the sittella request scoped auth.
type AuthContext interface {
	// Set adds the given data to the auth store and sets a session cookie.
	// May return backend specific errors or if the data is not serialisable by the configured codec.
	Set(data any, ttlOverride ...time.Duration) error

	// Get retrieves the data from the auth store if a session cookie exists.
	// Returns ErrAuthNotFound if an auth session is not found or is expired.
	// May return backend specific errors or if the data is not serialisable by the configured codec.
	Get(dest any, ttlOverride ...time.Duration) error

	// Delete removes the data from the auth store and revokes the session cookie if it exists.
	// May return backend specific errors.
	Delete() error
}

// Auth defines the sittella application scoped auth.
type Auth interface {
	// Returns a request scoped auth.
	WithContext(w http.ResponseWriter, r *http.Request) AuthContext

	// Stops any resources utilised by the auth.
	// Note: this does not remove any data.
	Stop() error
}
