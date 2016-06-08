package main

import (
	"io/ioutil"
	"net/http"

	"github.com/garyburd/redigo/redis"
)

func isNameTaken(w http.ResponseWriter, r *http.Request) error {
	conn := Pool.Get()
	defer conn.Close()

	username, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return ErrGeneric
	}

	taken, err := redis.Bool(conn.Do("HEXISTS", "webapp:users", string(username)))
	if err != nil {
		return ErrDB
	}

	takenTemp, err := redis.Bool(conn.Do("EXISTS", "webapp:temp:users:"+string(username)))
	if err != nil {
		return ErrDB
	}

	takenPending, err := redis.Bool(conn.Do("EXISTS", "webapp:users:pending:"+string(username)))
	if err != nil {
		return ErrDB
	}

	if taken || takenTemp || takenPending {
		w.Write([]byte("true"))
	} else {
		w.Write([]byte("false"))
	}

	return nil
}

func isNameTakenHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	if r.Method == "POST" {
		if err := isNameTaken(w, r); err != nil {
			return
		}
	} else {
		http.Error(w, "POST ONLY", http.StatusMethodNotAllowed)
	}
}
