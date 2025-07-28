package sittella

import (
	"errors"
	"fmt"
	"net/http"
	"sync"

	"github.com/dimmerz92/sittella/core"
	"github.com/dimmerz92/sittella/database"
	"github.com/dimmerz92/sittella/mailer"
	"github.com/dimmerz92/sittella/sessions"
)

type Config struct {
	DB           *database.Database
	SessionStore sessions.Store
	Mailer       mailer.Mailer
}

type app struct {
	middleware   []core.MiddlewareFunc
	onStart      func()
	onStop       func()
	errorHandler func(c core.Context, err error)
	server       *http.Server
	mux          *http.ServeMux
	db           *database.Database
	sessionStore sessions.Store
	mailer       mailer.Mailer
}

func New(config Config) core.App {
	if config.DB == nil {
		panic("App requires non-nil database")
	}
	if config.SessionStore == nil {
		panic("App requires non-nil session store")
	}
	if config.Mailer == nil {
		panic("App requires non-nil mailer")
	}

	return &app{
		server:       &http.Server{},
		mux:          &http.ServeMux{},
		db:           config.DB,
		sessionStore: config.SessionStore,
		mailer:       config.Mailer,
	}
}

// OnStart runs the given callback immediately before the sever starts.
func (a *app) OnStart(callback func()) {
	a.onStart = callback
}

// On stop runs the given callback immediately before the server stops.
func (a *app) OnStop(callback func()) {
	a.onStop = callback
}

func (a *app) OnHandlerError(callback func(c core.Context, err error)) {
	a.errorHandler = callback
}

func (a *app) serveMux() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handler := func(c core.Context) error {
			a.mux.ServeHTTP(c.Response(), c.Request())
			return nil
		}

		for i := len(a.middleware) - 1; i >= 0; i-- {
			handler = a.middleware[i](handler)
		}

		if err := handler(&context{res: w, req: r}); err != nil {
			panic(err)
		}
	})
}

// Start runs the http server on the given port number.
func (a *app) Start(port int) error {
	if port < 1 || port > 65535 {
		return errors.New("port must be between 1 and 65535")
	}
	if a.onStart != nil {
		a.onStart()
	}

	a.server = &http.Server{
		Handler: a.serveMux(),
		Addr:    fmt.Sprintf(":%d", port),
	}

	fmt.Printf("listening on port %d", port)
	return a.server.ListenAndServe()
}

// Use applies the given middleware to all registered handlers.
func (a *app) Use(middleware ...core.MiddlewareFunc)

func (a *app) serve(method, path string, handler core.HandlerFunc, middleware ...core.MiddlewareFunc) {
	for i := len(middleware) - 1; i >= 0; i-- {
		handler = middleware[i](handler)
	}

	a.mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != method || method != "any" {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		ctx := &context{
			req:     r,
			res:     w,
			store:   sync.Map{},
			db:      a.db,
			mailer:  a.mailer,
			session: a.sessionStore.Session(w, r),
		}

		if err := handler(ctx); err != nil {
			if a.errorHandler != nil {
				a.errorHandler(ctx, err)
				return
			}
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
	})
}

// Any registers a handler for any HTTP request method.
func (a *app) Any(path string, handler core.HandlerFunc, middleware ...core.MiddlewareFunc) {
	a.serve("any", path, handler, middleware...)
}

// GET registers a handler for HTTP GET requests.
func (a *app) GET(path string, handler core.HandlerFunc, middleware ...core.MiddlewareFunc) {
	a.serve(http.MethodGet, path, handler, middleware...)
}

// POST registers a handler for HTTP POST requests.
func (a *app) POST(path string, handler core.HandlerFunc, middleware ...core.MiddlewareFunc) {
	a.serve(http.MethodPost, path, handler, middleware...)
}

// PUT registers a handler for HTTP PUT requests.
func (a *app) PUT(path string, handler core.HandlerFunc, middleware ...core.MiddlewareFunc) {
	a.serve(http.MethodPut, path, handler, middleware...)
}

// DELETE registers a handler for HTTP DELETE requests.
func (a *app) DELETE(path string, handler core.HandlerFunc, middleware ...core.MiddlewareFunc) {
	a.serve(http.MethodDelete, path, handler, middleware...)
}

// HEAD registers a handler for HTTP HEAD requests.
func (a *app) HEAD(path string, handler core.HandlerFunc, middleware ...core.MiddlewareFunc) {
	a.serve(http.MethodHead, path, handler, middleware...)
}

// CONNECT registers a handler for HTTP CONNECT requests.
func (a *app) CONNECT(path string, handler core.HandlerFunc, middleware ...core.MiddlewareFunc) {
	a.serve(http.MethodConnect, path, handler, middleware...)
}

// OPTIONS registers a handler for HTTP OPTIONS requests.
func (a *app) OPTIONS(path string, handler core.HandlerFunc, middleware ...core.MiddlewareFunc) {
	a.serve(http.MethodOptions, path, handler, middleware...)
}

// TRACE registers a handler for HTTP TRACE requests.
func (a *app) TRACE(path string, handler core.HandlerFunc, middleware ...core.MiddlewareFunc) {
	a.serve(http.MethodTrace, path, handler, middleware...)
}

// PATCH registers a handler for HTTP PATCH requests.
func (a *app) PATCH(path string, handler core.HandlerFunc, middleware ...core.MiddlewareFunc) {
	a.serve(http.MethodPatch, path, handler, middleware...)
}
