package main

import (
	"html/template"
	"log"
	"net/http"

	"github.com/garyburd/redigo/redis"
)

func serveAdminPanel(w http.ResponseWriter, r *http.Request) error {
	conn := Pool.Get()
	defer conn.Close()

	users, err := redis.Strings(conn.Do("LRANGE", "webapp:users:pending", 0, -1))
	if err != nil {
		return ErrDB
	}

	t, err := template.ParseFiles("pages/admin.html")
	if err != nil {
		return err
	}
	t.Execute(w, users)

	return nil
}

func serveHome(w http.ResponseWriter, r *http.Request, user string) error {
	conn := Pool.Get()
	defer conn.Close()

	// Fetch the last ten message
	messages, err := redis.Strings(conn.Do("LRANGE", "webapp:messages:"+user, 0, 9))
	if err != nil {
		return ErrDB
	}

	t, err := template.ParseFiles("pages/home.html")
	if err != nil {
		return err
	}
	t.Execute(w, messages)

	return nil
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	if r.Method == "GET" {
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")
		w.Header().Set("Content-Type", "text/html")

		logged, user, err := isLoggedIn(w, r)
		if err != nil {
			return
		}
		if logged {
			admin, err := isUserAdmin(user)
			if err != nil {
				return
			}

			if admin {
				serveAdminPanel(w, r)
			} else if err := serveHome(w, r, user); err != nil {
				http.Error(w, "Internal error", http.StatusInternalServerError)
			}
		} else {
			t, err := template.ParseFiles("pages/signin.html")
			if err != nil {
				http.Error(w, "Internal error", http.StatusInternalServerError)
				log.Printf("handling %q: %v", r.RequestURI, err)
			}
			t.Execute(w, nil)
		}
	} else {
		http.Error(w, "GET ONLY", http.StatusMethodNotAllowed)
	}
}
