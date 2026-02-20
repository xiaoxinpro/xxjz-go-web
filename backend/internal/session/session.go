package session

import (
	"net/http"

	"github.com/gorilla/sessions"
)

const (
	SessionName     = "xxjz_session"
	KeyUID          = "uid"
	KeyUsername     = "username"
	KeyUserShell    = "user_shell"
	KeyWxOpenID     = "wx_openid"
	KeyWxSessionKey = "wx_session_key"
	KeyWxUnionID    = "wx_unionid"
)

// Store holds the session store (cookie-based).
var Store *sessions.CookieStore

// Init initializes the session store with a secret.
func Init(secret string) {
	Store = sessions.NewCookieStore([]byte(secret))
	Store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7, // 7 days
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
}

// SessionStore is implemented by gin-contrib/sessions.Session (Get/Set).
type SessionStore interface {
	Get(key interface{}) interface{}
	Set(key interface{}, val interface{})
}

// GetUID returns the logged-in user ID from session, or 0.
func GetUID(s SessionStore) int64 {
	v := s.Get(KeyUID)
	if v == nil {
		return 0
	}
	switch u := v.(type) {
	case int64:
		return u
	case int:
		return int64(u)
	case float64:
		return int64(u)
	}
	return 0
}

// GetUsername returns the logged-in username from session.
func GetUsername(s SessionStore) string {
	v := s.Get(KeyUsername)
	if v == nil {
		return ""
	}
	s2, _ := v.(string)
	return s2
}

// GetShell returns the session shell for validation.
func GetShell(s SessionStore) string {
	v := s.Get(KeyUserShell)
	if v == nil {
		return ""
	}
	s2, _ := v.(string)
	return s2
}

// SetLogin sets session values after successful login. userShell = md5(username+dbPassword).
func SetLogin(s SessionStore, uid int64, username, userShell string) {
	s.Set(KeyUID, uid)
	s.Set(KeyUsername, username)
	s.Set(KeyUserShell, userShell)
}

// Clear clears all session values (logout). For gorilla/sessions.Session.
func Clear(s *sessions.Session) {
	for k := range s.Values {
		delete(s.Values, k)
	}
}

// Options for cookie (used by main when creating store).
func Options() sessions.Options {
	return sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
}

var _ = http.NoBody // use net/http
