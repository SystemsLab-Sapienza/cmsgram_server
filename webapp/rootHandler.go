package main

import (
	"log"
	"net/http"

	"github.com/garyburd/redigo/redis"
)

func serveAdminPanel(w http.ResponseWriter, r *http.Request) error {
	type user struct {
		Username string
		Email    string
	}
	var data = struct {
		NewUsers []user
		Accounts []account
	}{}

	conn := pool.Get()
	defer conn.Close()

	// Get the list of usernames waiting to be accepted
	users, err := redis.Strings(conn.Do("LRANGE", "webapp:users:pending", 0, -1))
	if err != nil {
		return ErrDB
	}

	// For each username get the associated email and add both to the slice
	for _, u := range users {
		email, err := redis.String(conn.Do("HGET", "webapp:users:pending:"+u, "email"))
		if err != nil {
			return ErrDB
		}

		data.NewUsers = append(data.NewUsers, user{u, email})
	}

	data.Accounts, err = getAccounts()
	if err != nil {
		return err
	}

	templates.ExecuteTemplate(w, "admin.html", data)

	return nil
}

func serveHome(w http.ResponseWriter, r *http.Request, user string) error {
	conn := pool.Get()
	defer conn.Close()

	var messages = make([]string, 0)

	// Fetch the ids of last ten messages
	ids, err := redis.Strings(conn.Do("LRANGE", "webapp:users:messages:"+user, 0, 9))
	if err != nil {
		return ErrDB
	}

	for _, i := range ids {
		msg, err := redis.String(conn.Do("HGET", "webapp:messages:"+i, "content"))
		if err != nil {
			return err
		}
		messages = append(messages, msg)
	}

	templates.ExecuteTemplate(w, "home.html", messages)

	return nil
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	switch r.Method {
	case "GET":
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")
		w.Header().Set("Content-Type", "text/html")

		logged, user, err := isLoggedIn(w, r)
		if err != nil {
			return
		}
		if !logged {
			templates.ExecuteTemplate(w, "signin.html", nil)
			return
		}

		admin, err := isUserAdmin(user)
		if err != nil {
			http.Error(w, "Internal error", http.StatusInternalServerError)
			log.Printf("handling %q: %v", r.RequestURI, err)
			return
		}

		if admin {
			if err := serveAdminPanel(w, r); err != nil {
				http.Error(w, "Internal error", http.StatusInternalServerError)
				log.Printf("handling %q: %v", r.RequestURI, err)
			}
		} else {
			if err := serveHome(w, r, user); err != nil {
				http.Error(w, "Internal error", http.StatusInternalServerError)
				log.Printf("handling %q: %v", r.RequestURI, err)
			}
		}
	default:
		http.Error(w, "GET ONLY", http.StatusMethodNotAllowed)
	}
}
