package core

import (
	"net/http"

	"github.com/a-h/templ"
	"github.com/dimmerz92/sittella/database"
	"github.com/dimmerz92/sittella/mailer"
	"github.com/dimmerz92/sittella/sessions"
)

// Context defines the request scoped context.
type Context interface {
	// Request returns the underlying request.
	Request() *http.Request

	// Response returns the underlying response writer.
	Response() http.ResponseWriter

	// Set adds the key value pair to the context store.
	Set(key string, value any)

	// Get returns the value mapped to by the given key from the context store if
	// it exists.
	Get(key string) (any, bool)

	// DB returns the underlying database.
	DB() *database.Database

	// Mailer returns the underlying mailer.
	Mailer() mailer.Mailer

	// Session returns the unique user session.
	Session() sessions.Session

	// HTML writes the given status and html string to the response.
	HTML(status int, html string) error

	// JSON writes the given status and json data to the response.
	JSON(status int, data any) error

	// String writes the given status and string to the response.
	String(status int, text string) error

	// NoContent writes the given status to the response without a body.
	NoContent(status int) error

	// NotFound writes a status 404 to the response without a body.
	NotFound() error

	// Render writes the given status and templates to the response.
	Render(status int, tpls ...templ.Component) error

	// Redirect is a HTMX aware redirect method.
	// Non-HTMX requests result in a redirect to the given path with status.
	// HTMX requests return a 200 - OK status with HTMX redirect headers.
	// https://github.com/bigskysoftware/htmx/issues/2052#issuecomment-1979805051
	Redirect(status int, path string) error

	// IsHTMX returns true if the current request is HTMX, otherwise false.
	IsHTMX() bool
}
