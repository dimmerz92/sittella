package sittella

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/a-h/templ"
	"github.com/dimmerz92/sittella/auth"
	"github.com/dimmerz92/sittella/database"
	"github.com/dimmerz92/sittella/mailer"
	"github.com/dimmerz92/sittella/sessions"
)

type Context[Queries any] struct {
	res      http.ResponseWriter
	req      *http.Request
	store    sync.Map
	db       database.Database[Queries]
	mailer   mailer.Mailer
	auth     auth.AuthContext
	sessions sessions.SessionStoreContext
}

// Response returns the underlying response writer.
func (c *Context[Queries]) Response() http.ResponseWriter {
	return c.res
}

// Request returns the underlying request.
func (c *Context[Queries]) Request() *http.Request {
	return c.req
}

// Set adds the data mapped to by key in the request scoped store.
// Data stored using this method will not be persisted past the end of the current request.
// For a more persistent store, use the Sessions() method.
func (c *Context[Queries]) Set(key string, data any) {
	c.store.Store(key, data)
}

// Get retrieves the data mapped to by key from the request scoped store if it exists.
func (c *Context[Queries]) Get(key string) (any, bool) {
	return c.store.Load(key)
}

// DB returns access to the underlying database and user-defined queries.
func (c *Context[Queries]) DB() database.Database[Queries] {
	return c.db
}

func (c *Context[Queries]) Mailer() mailer.Mailer {
	return c.mailer
}

// Auth returns access to the request scoped auth store.
func (c *Context[Queries]) Auth() auth.AuthContext {
	return c.auth
}

// Sessions returns access to the request scoped session store.
func (c *Context[Queries]) Sessions() sessions.SessionStoreContext {
	return c.sessions
}

// NoContent writes the given status with no body to the response writer.
func (c *Context[Queries]) NoContent(status int) error {
	c.res.WriteHeader(status)
	return nil
}

// NotFound writes a 404 status with no body to the response writer.
func (c *Context[Queries]) NotFound() error {
	c.res.WriteHeader(http.StatusNotFound)
	return nil
}

func (c *Context[Queries]) respond(status int, content []byte, contentType string) error {
	c.res.Header().Set("Content-Type", contentType)
	c.res.WriteHeader(status)
	_, err := c.res.Write(content)
	return err
}

// String writes the given status and text to the response writer as plain text.
func (c *Context[Queries]) String(status int, text string) error {
	return c.respond(status, []byte(text), "text/plain")
}

// HTML writes the given status and html string to the response writer as HTML.
func (c *Context[Queries]) HTML(status int, html string) error {
	return c.respond(status, []byte(html), "text/html")
}

// JSON writes the given status and data to the response writer as JSON.
func (c *Context[Queries]) JSON(status int, data any) error {
	encoded, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return c.respond(status, encoded, "application/json")
}

// Render writes the given status and any number of templates to the response as HTML.
func (c *Context[Queries]) Render(status int, tpls ...templ.Component) error {
	buf := templ.GetBuffer()
	defer templ.ReleaseBuffer(buf)

	for _, tpl := range tpls {
		if err := tpl.Render(c.req.Context(), buf); err != nil {
			return err
		}
	}

	return c.HTML(status, buf.String())
}

// IsHTMX returns true if a request was made by HTMX, otherwise false.
func (c *Context[Queries]) IsHTMX() bool {
	return c.req.Header.Get("Hx-Request") == "true"
}

// Redirect is a HTMX aware redirect method.
// A standard http redirect is made with the given status if the request was not made by HTMX.
// Otherwise, the Hx-Redirect header is set to path and a status 200 is returned instead of the given status.
// See: https://github.com/bigskysoftware/htmx/issues/2052#issuecomment-1979805051
func (c *Context[Queries]) Redirect(status int, path string) error {
	if c.IsHTMX() {
		c.res.Header().Set("Hx-Redirect", path)
		c.NoContent(http.StatusOK)
		return nil
	}
	http.Redirect(c.res, c.req, path, status)
	return nil
}
