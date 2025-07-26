package sessions

import (
	"net/http"
	"time"
)

var DefaultCookie = CookieOpts{
	Name:     "session",
	Path:     "/",
	HttpOnly: true,
	Secure:   true,
	SameSite: http.SameSiteStrictMode,
}

type CookieOpts struct {
	// Name sets the cookie name.
	Name string

	// Domain sets the cookie's domain origin.
	Domain string

	// Path specifies the path the cookie applies to.
	Path string

	// HttpOnly specifies if the cookie is available to JavaScript.
	HttpOnly bool

	// Secure specifies if the cookie is served over HTTPS only.
	Secure bool

	// Partitioned specifies if the cookie is partitioned.
	Partitioned bool

	// SameSite specifies the same site settings for the cookie.
	SameSite http.SameSite
}

func (c *CookieOpts) ToCookie(value string, ttl time.Duration) *http.Cookie {
	return &http.Cookie{
		Name:        c.Name,
		Domain:      c.Domain,
		Path:        c.Path,
		Value:       value,
		Expires:     time.Now().Add(ttl),
		MaxAge:      int(ttl.Seconds()),
		HttpOnly:    c.HttpOnly,
		Secure:      c.Secure,
		Partitioned: c.Partitioned,
		SameSite:    c.SameSite,
	}
}

func (c *CookieOpts) ToRevoked() *http.Cookie {
	return &http.Cookie{
		Name:        c.Name,
		Domain:      c.Domain,
		Path:        c.Path,
		Value:       "",
		Expires:     time.Unix(0, 0),
		MaxAge:      -1,
		HttpOnly:    c.HttpOnly,
		Secure:      c.Secure,
		Partitioned: c.Partitioned,
		SameSite:    c.SameSite,
	}
}
