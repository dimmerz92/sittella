package auth

import (
	"net/http"
	"time"

	"github.com/dimmerz92/sittella/utils"
)

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

// SetSessionCookie adds a session cookie to the response writer.
func SetSessionCookie(w http.ResponseWriter, sessionId string, ttl time.Duration, cookieOpts CookieOpts) {
	http.SetCookie(w, &http.Cookie{
		Name:     utils.Coalesce(cookieOpts.Name, DefaultCookieOpts.Name),
		Path:     utils.Coalesce(cookieOpts.Path, DefaultCookieOpts.Path),
		Value:    sessionId,
		Expires:  time.Now().Add(ttl), // for compatibility
		MaxAge:   int(ttl.Seconds()),
		HttpOnly: cookieOpts.HttpOnly != "false",
		Secure:   cookieOpts.Secure != "false",
		SameSite: utils.Coalesce(cookieOpts.SameSite, DefaultCookieOpts.SameSite),
	})
}

// RevokeSessionCookie invalidates a session cookie on the response writer.
func RevokeSessionCookie(w http.ResponseWriter, cookieOpts CookieOpts) {
	http.SetCookie(w, &http.Cookie{
		Name:     utils.Coalesce(cookieOpts.Name, DefaultCookieOpts.Name),
		Path:     utils.Coalesce(cookieOpts.Path, DefaultCookieOpts.Path),
		Value:    "",
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		HttpOnly: cookieOpts.HttpOnly != "false",
		Secure:   cookieOpts.Secure != "false",
		SameSite: utils.Coalesce(cookieOpts.SameSite, DefaultCookieOpts.SameSite),
	})
}

// GetSessionIdFromCookie returns the session ID if it exists or an empty string.
func GetSessionIdFromCookie(r *http.Request, cookieOpts CookieOpts) (string, bool) {
	cookie, _ := r.Cookie(utils.Coalesce(cookieOpts.Name, DefaultCookieOpts.Name))
	if cookie != nil {
		return cookie.Value, true
	}
	return "", false
}
