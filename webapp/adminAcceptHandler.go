package main

import (
	"log"
	"net/http"
	"strconv"

	"github.com/garyburd/redigo/redis"
)

func accountAccept(w http.ResponseWriter, r *http.Request) error {
	var (
		userData = struct {
			Email    string `redis:"email"`
			Username string `redis:"username"`
			Hash     string `redis:"hash"`
			Auth     string `redis:"auth"`
		}{}
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

	admin, err := isUserAdmin(uid)
	if err != nil {
		return err
	} else if !admin {
		http.NotFound(w, r)
		return nil
	}

	if len(username) == 0 {
		return ErrFieldEmpty
	}

	conn := pool.Get()
	defer conn.Close()

	data, err := redis.Values(conn.Do("HGETALL", "webapp:users:pending:"+username))
	if err != nil {
		return ErrDB
	}
	if data == nil {
		return ErrGeneric
	}

	err = redis.ScanStruct(data, &userData)
	if err != nil {
		return ErrDB
	}

	user, err := redis.Int(conn.Do("INCR", "webapp:users:counter"))
	if err != nil {
		return ErrDB
	}

	conn.Send("MULTI")
	conn.Send("HSET", "webapp:users", userData.Username, user)
	conn.Send("HSET", "webapp:users:email", userData.Email, strconv.Itoa(user))
	conn.Send("LREM", "webapp:users:pending", 0, userData.Username)
	conn.Send("RENAME", "webapp:users:pending:"+username, "webapp:users:"+strconv.Itoa(user))
	_, err = conn.Do("EXEC")
	if err != nil {
		return ErrDB
	}

	// Send email update to user
	subject := "Account attivato"
	body := "Il tuo account è stato approvato dall'amministratore ed è ora attivo:\n" +
		config.Domain
	go sendEmail(userData.Email, subject, body)

	http.Redirect(w, r, "/", http.StatusSeeOther)
	return nil
}

func adminAcceptHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	switch r.Method {
	case "POST":
		if err := accountAccept(w, r); err != nil {
			log.Println(err)
			return
		}
	default:
		http.Error(w, "POST ONLY", http.StatusMethodNotAllowed)
	}
}
