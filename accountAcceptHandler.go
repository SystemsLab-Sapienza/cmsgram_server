package main

import (
	"log"
	"net/http"
	"strconv"

	"github.com/garyburd/redigo/redis"
)

func accountAccept(w http.ResponseWriter, r *http.Request) error {
	var (
		renderUserData = struct {
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

	data, err := redis.Values(conn.Do("HGETALL", "webapp:users:pending:"+username))
	if err != nil {
		return ErrDB
	}
	if data == nil {
		return ErrGeneric // TODO use more specific error
	}

	err = redis.ScanStruct(data, &renderUserData)
	if err != nil {
		return ErrDB
	}

	user, err := redis.Int(conn.Do("INCR", "webapp:users:counter"))
	if err != nil {
		return ErrDB
	}

	conn.Send("MULTI")
	conn.Send("HSET", "webapp:users", renderUserData.Username, user)
	conn.Send("SADD", "webapp:users:email", renderUserData.Email)
	conn.Send("LREM", "webapp:users:pending", 0, renderUserData.Username)
	conn.Send("RENAME", "webapp:users:pending:"+username, "webapp:users:"+strconv.Itoa(user))
	_, err = conn.Do("EXEC")
	if err != nil {
		return ErrDB
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
	return nil
}

func accountAcceptHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	if r.Method == "POST" {
		if err := accountAccept(w, r); err != nil {
			errmsg := err.Error()
			log.Println(errmsg)
			return
		}
	} else {
		http.Error(w, "POST ONLY", http.StatusMethodNotAllowed)
	}
}
