package main

import (
	"net/http"
	// "strconv"

	"github.com/garyburd/redigo/redis"
)

func accountActivate(w http.ResponseWriter, r *http.Request) error {
	var (
		userData = struct {
			Email    string `redis:"email"`
			Username string `redis:"username"`
			Hash     string `redis:"hash"`
			Auth     string `redis:"auth"`
		}{}
		token = r.FormValue("token")
	)

	conn := Pool.Get()
	defer conn.Close()

	user, err := redis.Values(conn.Do("HGETALL", "webapp:temp:token:"+token))
	if err != nil {
		return ErrDB
	}
	if user == nil {
		return ErrBadToken
	}

	err = redis.ScanStruct(user, &userData)
	if err != nil {
		return ErrDB
	}

	_, err = conn.Do("HMSET", redis.Args{}.Add("webapp:users:pending:"+userData.Username).AddFlat(&userData)...)
	if err != nil {
		return ErrDB
	}

	conn.Send("MULTI")
	conn.Send("DEL", "webapp:temp:token:"+token)
	conn.Send("DEL", "webapp:temp:users:"+userData.Username)
	conn.Send("DEL", "webapp:temp:email:"+userData.Email)
	conn.Send("RPUSH", "webapp:users:pending", userData.Username)
	// HMSET
	_, err = conn.Do("EXEC")
	if err != nil {
		return ErrDB
	}

	return nil
}

func accountActivateHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	if r.Method == "GET" {
		if err := accountActivate(w, r); err != nil {
			w.Write([]byte(err.Error()))
			return
		} else {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
	} else {
		http.Error(w, "GET ONLY", http.StatusMethodNotAllowed)
		return
	}
}
