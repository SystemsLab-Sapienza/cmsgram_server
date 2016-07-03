package main

import (
	"log"
	"net/http"

	"github.com/garyburd/redigo/redis"
)

func accountDeny(w http.ResponseWriter, r *http.Request) error {
	var (
		username = r.PostFormValue("username")
	)

	logged, uid, err := isLoggedIn(w, r)
	if err != nil {
		return ErrGeneric
	}
	if !logged {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return nil
	}

	conn := Pool.Get()
	defer conn.Close()

	admin, err := isUserAdmin(uid)
	if err != nil {
		return err
	} else if !admin {
		http.NotFound(w, r)
		return nil
	}

	if username == "" {
		return ErrFieldEmpty
	}

	// Get user's email address
	to, err := redis.String(conn.Do("HGET", "webapp:users:pending:"+username, "email"))
	if err != nil {
		return ErrDB
	}

	// Delete account data
	conn.Send("MULTI")
	conn.Send("LREM", "webapp:users:pending", 0, username)
	conn.Send("DEL", "webapp:users:pending:"+username)
	_, err = conn.Do("EXEC")
	if err != nil {
		return ErrDB
	}

	// Send email update to user
	subject := "Attivazione account rifiutata"
	body := "L'amministratore ha rifiutato la tua richiesta di attivazione dell'account."
	go sendEmail(to, subject, body)

	http.Redirect(w, r, "/", http.StatusSeeOther)
	return nil
}

func adminDenyHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	if r.Method == "POST" {
		if err := accountDeny(w, r); err != nil {
			errmsg := err.Error()
			log.Println(errmsg)
			return
		}
	} else {
		http.Error(w, "POST ONLY", http.StatusMethodNotAllowed)
	}
}
