package sittella

import (
	"encoding/json"
	"net/http"

	"github.com/a-h/templ"
	"github.com/dimmerz92/sittella/auth"
	"github.com/dimmerz92/sittella/database"
	"github.com/dimmerz92/sittella/mailer"
	"github.com/dimmerz92/sittella/sessions"
)

type Context[Queries any] struct {
	res      http.ResponseWriter
	req      *http.Request
	db       database.Database[Queries]
	mailer   mailer.Mailer
	auth     auth.AuthContext
	sessions sessions.SessionStoreContext
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
