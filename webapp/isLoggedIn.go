package main

import (
	"net/http"

	"github.com/garyburd/redigo/redis"
)

// Returns true if user is logged in and the uid associated with it.
// Otherwise it returns false, ""
func isLoggedIn(w http.ResponseWriter, r *http.Request) (bool, string, error) {
	// No cookie
	cookie, err := r.Cookie("auth")
	if err != nil {
		return false, "", nil
	}

	conn := pool.Get()
	defer conn.Close()

	// Nonexistent session
	uid, err := redis.String(conn.Do("GET", "webapp:users:auth:session:"+cookie.Value))
	if uid == "" {
		return false, "", nil
	} else if err != nil {
		return false, "", ErrDB
	}

	// Expired session
	auth, err := redis.String(conn.Do("HGET", "webapp:users:"+uid, "auth"))
	if err != nil {
		return false, "", ErrDB
	} else if cookie.Value != auth {
		// Delete session
		_, err := conn.Do("DEL", "webapp:users:auth:session:"+cookie.Value)
		if err != nil {
			return false, "", ErrDB
		}

		// Delete cookie
		newCookie := http.Cookie{Name: "auth", Value: "", MaxAge: -1}
		http.SetCookie(w, &newCookie)

		return false, "", nil
	}

	return true, uid, nil
}
