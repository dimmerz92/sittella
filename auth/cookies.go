package auth

import "net/http"

// CookieOpts sets the cookie parameters for the Auth session cookies.
type CookieOpts struct {
	// Name specifies the cookie name. Defaults to "__sittella_lab".
	Name string

	// Path specifies the cookie path. Defaults to "/".
	Path string

	// HttpOnly specifies if the cookie HTTP only. Defaults to "true".
	// Note: any string that is not "false" (case sensitive) is considered as "true".
	HttpOnly string

	// Secure specifies whether the cookie is served over HTTPS.
	// Note: any string that is not "false" (case sensitive) is considered as "true".
	Secure string

	// SameSite specifies cross-origin policy of the cookie. Defaults to strict.
	SameSite http.SameSite
}

var DefaultCookieOpts = CookieOpts{
	Name:     "__sittella_auth",
	Path:     "/",
	HttpOnly: "true",
	Secure:   "true",
	SameSite: http.SameSiteStrictMode,
}
