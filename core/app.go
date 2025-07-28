package core

type HandlerFunc func(c Context) error
type MiddlewareFunc func(next HandlerFunc) HandlerFunc

// App defines the interface for the central application.
type App interface {
	// OnStart runs the given callback immediately before the sever starts.
	OnStart(callback func())

	// On stop runs the given callback immediately before the server stops.
	OnStop(callback func())

	// OnHandlerError runs the given callback as a central error handler.
	OnHandlerError(callback func(c Context, err error))

	// Start runs the http server on the given port number.
	Start(port int) error

	// Use applies the given middleware to all registered handlers.
	Use(middleware ...MiddlewareFunc)

	// Any registers a handler for any HTTP request method.
	Any(path string, handler HandlerFunc, middleware ...MiddlewareFunc)

	// GET registers a handler for HTTP GET requests.
	GET(path string, handler HandlerFunc, middleware ...MiddlewareFunc)

	// POST registers a handler for HTTP POST requests.
	POST(path string, handler HandlerFunc, middleware ...MiddlewareFunc)

	// PUT registers a handler for HTTP PUT requests.
	PUT(path string, handler HandlerFunc, middleware ...MiddlewareFunc)

	// DELETE registers a handler for HTTP DELETE requests.
	DELETE(path string, handler HandlerFunc, middleware ...MiddlewareFunc)

	// HEAD registers a handler for HTTP HEAD requests.
	HEAD(path string, handler HandlerFunc, middleware ...MiddlewareFunc)

	// OPTIONS registers a handler for HTTP OPTIONS requests.
	OPTIONS(path string, handler HandlerFunc, middleware ...MiddlewareFunc)

	// CONNECT registers a handler for HTTP CONNECT requests.
	CONNECT(path string, handler HandlerFunc, middleware ...MiddlewareFunc)

	// TRACE registers a handler for HTTP TRACE requests.
	TRACE(path string, handler HandlerFunc, middleware ...MiddlewareFunc)

	// PATCH registers a handler for HTTP PATCH requests.
	PATCH(path string, handler HandlerFunc, middleware ...MiddlewareFunc)
}
