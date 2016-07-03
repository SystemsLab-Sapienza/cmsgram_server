package main

import (
	"html/template"
	"log"
	"net/http"
	"strings"

	"github.com/garyburd/redigo/redis"
)

func accountDelete(w http.ResponseWriter, r *http.Request) error {
	var user = struct {
		Email    string `redis:"email"`
		Username string `redis:"username"`
		Auth     string `redis:"auth"`
	}{}

	conn := Pool.Get()
	defer conn.Close()

	logged, uid, err := isLoggedIn(w, r)
	if err != nil {
		return ErrGeneric
	}
	if !logged {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return nil
	}

	data, err := redis.Values(conn.Do("HGETALL", "webapp:users:"+uid))
	if err != nil {
		return ErrDB
	}

	err = redis.ScanStruct(data, &user)
	if err != nil {
		return ErrDB
	}

	lname, err := redis.String(conn.Do("HGET", "webapp:users:data:"+uid, "cognome"))
	if err != nil && err != redis.ErrNil {
		return ErrDB
	}

	conn.Send("MULTI")
	conn.Send("HDEL", "webapp:users", user.Username)
	conn.Send("SREM", "webapp:users:email", user.Email)
	conn.Send("DEL", "webapp:users:auth:session:"+user.Auth)
	conn.Send("DEL", "webapp:users:"+uid)
	conn.Send("DEL", "webapp:users:data:"+uid)
	conn.Send("DEL", "webapp:users:data:email:"+uid)
	conn.Send("DEL", "webapp:users:data:url:"+uid)
	conn.Send("DEL", "webapp:messages:"+uid)
	if lname != "" {
		conn.Send("SREM", "webapp:users:info:"+strings.ToLower(lname), uid)
	}
	_, err = conn.Do("EXEC")
	if err != nil {
		return ErrDB
	}

	t, err := template.ParseFiles("templates/confirm.html")
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		log.Printf("handling %q: %v", r.RequestURI, err)
		return err
	}
	t.Execute(w, "L'account è stato eliminato.")

	return nil
}

func accountDeleteHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	if r.Method == "POST" {
		if err := accountDelete(w, r); err != nil {
			http.Error(w, "Internal error", http.StatusInternalServerError)
			log.Printf("handling %q: %v", r.RequestURI, err)
			return
		}
	} else {
		http.Error(w, "POST ONLY", http.StatusMethodNotAllowed)
	}
}
