package main

import (
	"net/http"
	// "github.com/garyburd/redigo/redis"
)

func signout(w http.ResponseWriter, r *http.Request) error {
	conn := Pool.Get()
	defer conn.Close()

	logged, uid, err := isLoggedIn(w, r)
	if err != nil {
		return err
	}
	if logged {
		cookie, _ := r.Cookie("auth")

		_, err := conn.Do("HSET", "webapp:users:"+uid, "auth", "")
		if err != nil {
			return ErrDB
		}

		_, err = conn.Do("DEL", "webapp:users:auth:session:"+cookie.Value)
		if err != nil {
			return ErrDB
		}

		// Delete cookie
		newCookie := http.Cookie{Name: "auth", Value: "", MaxAge: -1}
		http.SetCookie(w, &newCookie)
	}

	return nil
}

func signoutHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	if r.Method == "GET" {
		if err := signout(w, r); err != nil {
			return
		}
		http.Redirect(w, r, "/", http.StatusSeeOther)
	} else {
		http.Error(w, "GET ONLY", http.StatusMethodNotAllowed)
	}
}
