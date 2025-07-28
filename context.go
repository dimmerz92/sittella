package sittella

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/a-h/templ"
	"github.com/dimmerz92/sittella/database"
	"github.com/dimmerz92/sittella/mailer"
	"github.com/dimmerz92/sittella/sessions"
)

type context struct {
	req     *http.Request
	res     http.ResponseWriter
	store   sync.Map
	db      *database.Database
	mailer  mailer.Mailer
	session sessions.Session
}

// Request returns the underlying request.
func (c *context) Request() *http.Request { return c.req }

// Response returns the underlying response writer.
func (c *context) Response() http.ResponseWriter { return c.res }

// Set adds the key value pair to the context store.
func (c *context) Set(key string, value any) { c.store.Store(key, value) }

// Get returns the value mapped to by the given key from the context store if
// it exists.
func (c *context) Get(key string) (any, bool) { return c.store.Load(key) }

// DB returns the underlying database.
func (c *context) DB() *database.Database { return c.db }

// Mailer returns the underlying mailer.
func (c *context) Mailer() mailer.Mailer { return c.mailer }

// Session returns the unique user session.
func (c *context) Session() sessions.Session { return c.session }

// HTML writes the given status and html string to the response.
func (c *context) HTML(status int, html string) error {
	c.res.Header().Set("Content-Type", "text/html")
	c.res.WriteHeader(status)
	_, err := c.res.Write([]byte(html))
	return err
}

// JSON writes the given status and json data to the response.
func (c *context) JSON(status int, data any) error {
	encoded, err := json.Marshal(data)
	if err == nil {
		c.res.Header().Set("Content-Type", "application/json")
		c.res.WriteHeader(status)
		_, err = c.res.Write(encoded)
	}
	return err
}

// String writes the given status and string to the response.
func (c *context) String(status int, text string) error {
	c.res.Header().Set("Content-Type", "text/plain")
	c.res.WriteHeader(status)
	_, err := c.res.Write([]byte(text))
	return err
}

// NoContent writes the given status to the response without a body.
func (c *context) NoContent(status int) error {
	c.res.WriteHeader(status)
	return nil
}

// NotFound writes a status 404 to the response without a body.
func (c *context) NotFound() error {
	c.res.WriteHeader(http.StatusNotFound)
	return nil
}

// Render writes the given status and templates to the response.
func (c *context) Render(status int, tpls ...templ.Component) error {
	buf := templ.GetBuffer()
	defer templ.ReleaseBuffer(buf)

	for _, tpl := range tpls {
		if err := tpl.Render(c.req.Context(), buf); err != nil {
			return err
		}
	}

	return c.HTML(status, buf.String())
}

// Redirect is a HTMX aware redirect method.
// Non-HTMX requests result in a redirect to the given path with status.
// HTMX requests return a 200 - OK status with HTMX redirect headers.
// https://github.com/bigskysoftware/htmx/issues/2052#issuecomment-1979805051
func (c *context) Redirect(status int, path string) error {
	if c.IsHTMX() {
		c.res.Header().Set("Hx-Redirect", path)
		c.res.WriteHeader(http.StatusOK)
		return nil
	}

	http.Redirect(c.res, c.req, path, status)
	return nil
}

// IsHTMX returns true if the current request is HTMX, otherwise false.
func (c *context) IsHTMX() bool { return c.req.Header.Get("Hx-Request") == "true" }
