package auth

import (
	"net/http"
	"time"
)

func newCookie(name, value string) http.Cookie {
	year := time.Now().AddDate(1, 0, 0)
	cookie := http.Cookie{Name: name, Value: value, Expires: year, MaxAge: 3600 * 24 * 365}

	return cookie
}

func NewAuthCookie() http.Cookie {
	auth := NewBase36(32)
	cookie := newCookie("auth", auth)

	return cookie
}
