package sittella

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"

	"github.com/dimmerz92/sittella/auth"
	"github.com/dimmerz92/sittella/database"
	"github.com/dimmerz92/sittella/mailer"
	"github.com/dimmerz92/sittella/sessions"
)

// HandlerFunc defines the signature for request handlers.
type HandlerFunc[Queries any] func(*Context[Queries]) error

// MiddlewareFunc defines the signature for middleware functions.
type MiddlewareFunc[Queries any] func(HandlerFunc[Queries]) HandlerFunc[Queries]

// Config defines the configuration settings for the App.
type Config[Queries any] struct {
	Database     database.Database[Queries]
	Mailer       mailer.Mailer
	AuthStore    auth.Auth
	SessionStore sessions.SessionStore
	BeforeStart  func()
	BeforeStop   func()
}

type App[Queries any] struct {
	middleware  []MiddlewareFunc[Queries]
	mux         *http.ServeMux
	server      *http.Server
	db          database.Database[Queries]
	mailer      mailer.Mailer
	auth        auth.Auth
	sessions    sessions.SessionStore
	beforeStart func()
	beforeStop  func()
}

// New returns a new App instance.
func New[Queries any](config Config[Queries]) *App[Queries] {
	if config.Database == nil {
		panic("New App: database required")
	}

	if config.Mailer == nil {
		panic("New App: mailer required")
	}

	if config.AuthStore == nil {
		panic("New App: auth store required")
	}

	if config.SessionStore == nil {
		panic("New App: session store required")
	}

	if config.BeforeStart != nil {
		config.BeforeStart()
	}

	return &App[Queries]{
		mux:      http.NewServeMux(),
		db:       config.Database,
		mailer:   config.Mailer,
		sessions: config.SessionStore,
		auth:     config.AuthStore,
	}
}

func (a *App[Queries]) serveMux() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handler := func(c *Context[Queries]) error {
			a.mux.ServeHTTP(c.res, c.req)
			return nil
		}

		for i := len(a.middleware) - 1; i >= 0; i-- {
			handler = a.middleware[i](handler)
		}

		if err := handler(&Context[Queries]{res: w, req: r}); err != nil {
			panic(err)
		}
	})
}

// Use registers middleware to be applied globally to all routes in FIFO order.
func (a *App[Queries]) Use(middleware ...MiddlewareFunc[Queries]) {
	a.middleware = append(a.middleware, middleware...)
}

// Start runs the registered BeforeStart callback if it exists and starts the server on the given port.
func (a *App[Queries]) Start(port int) error {
	if port < 1 || port > 65535 {
		return errors.New("port must be between 1 and 65535")
	}
	if a.beforeStart != nil {
		a.beforeStart()
	}

	a.server = &http.Server{
		Handler: a.serveMux(),
		Addr:    fmt.Sprintf(":%d", port),
	}

	fmt.Printf("listening on port %d", port)
	return a.server.ListenAndServe()
}

// Stop runs the registered BeforeStop callback if it exists and shuts down the server gracefully.
func (a *App[Queries]) Stop(ctx context.Context) error {
	if a.beforeStop != nil {
		a.beforeStop()
	}

	return a.server.Shutdown(ctx)

}

func (a *App[Queries]) serve(method, path string, handler HandlerFunc[Queries], middleware ...MiddlewareFunc[Queries]) {
	for i := len(middleware) - 1; i >= 0; i-- {
		handler = middleware[i](handler)
	}

	a.mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != method {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		ctx := &Context[Queries]{
			res:      w,
			req:      r,
			store:    sync.Map{},
			db:       a.db,
			mailer:   a.mailer,
			auth:     a.auth.WithContext(w, r),
			sessions: a.sessions.WithContext(w, r),
		}

		if err := handler(ctx); err != nil {
			// TODO: add central error handling
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
	})
}

// GET registers a handler for HTTP GET requests.
func (a *App[Queries]) GET(path string, handler HandlerFunc[Queries], middleware ...MiddlewareFunc[Queries]) {
	a.serve(http.MethodGet, path, handler, middleware...)
}

// POST registers a handler for HTTP POST requests.
func (a *App[Queries]) POST(path string, handler HandlerFunc[Queries], middleware ...MiddlewareFunc[Queries]) {
	a.serve(http.MethodPost, path, handler, middleware...)
}

// PUT registers a handler for HTTP PUT requests.
func (a *App[Queries]) PUT(path string, handler HandlerFunc[Queries], middleware ...MiddlewareFunc[Queries]) {
	a.serve(http.MethodPut, path, handler, middleware...)
}

// DELETE registers a handler for HTTP DELETE requests.
func (a *App[Queries]) DELETE(path string, handler HandlerFunc[Queries], middleware ...MiddlewareFunc[Queries]) {
	a.serve(http.MethodDelete, path, handler, middleware...)
}

// HEAD registers a handler for HTTP HEAD requests.
func (a *App[Queries]) HEAD(path string, handler HandlerFunc[Queries], middleware ...MiddlewareFunc[Queries]) {
	a.serve(http.MethodHead, path, handler, middleware...)
}

// CONNECT registers a handler for HTTP CONNECT requests.
func (a *App[Queries]) CONNECT(path string, handler HandlerFunc[Queries], middleware ...MiddlewareFunc[Queries]) {
	a.serve(http.MethodConnect, path, handler, middleware...)
}

// OPTIONS registers a handler for HTTP OPTIONS requests.
func (a *App[Queries]) OPTIONS(path string, handler HandlerFunc[Queries], middleware ...MiddlewareFunc[Queries]) {
	a.serve(http.MethodOptions, path, handler, middleware...)
}

// TRACE registers a handler for HTTP TRACE requests.
func (a *App[Queries]) TRACE(path string, handler HandlerFunc[Queries], middleware ...MiddlewareFunc[Queries]) {
	a.serve(http.MethodTrace, path, handler, middleware...)
}

// PATCH registers a handler for HTTP PATCH requests.
func (a *App[Queries]) PATCH(path string, handler HandlerFunc[Queries], middleware ...MiddlewareFunc[Queries]) {
	a.serve(http.MethodPatch, path, handler, middleware...)
}
