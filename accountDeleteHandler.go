package main

import (
	"net/http"

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

	conn.Send("MULTI")
	conn.Send("HDEL", "webapp:users", user.Username)
	conn.Send("SREM", "webapp:users:email", user.Email)
	conn.Send("DEL", "webapp:users:auth:"+user.Auth)
	conn.Send("DEL", "webapp:users:"+uid)
	conn.Send("DEL", "webapp:users:data:"+uid)
	conn.Send("DEL", "webapp:users:data:email:"+uid)
	conn.Send("DEL", "webapp:users:data:url:"+uid)
	conn.Send("DEL", "webapp:messages:"+uid)
	_, err = conn.Do("EXEC")
	if err != nil {
		return ErrDB
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)

	return nil
}

func accountDeleteHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	if r.Method == "POST" {
		accountDelete(w, r)
	} else {
		http.Error(w, "POST ONLY", http.StatusMethodNotAllowed)
	}
}
