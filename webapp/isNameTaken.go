package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/garyburd/redigo/redis"
)

func isNameTaken(w http.ResponseWriter, r *http.Request) error {
	var payload = struct {
		Key   string
		Value string
	}{}

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return ErrGeneric
	}

	err = json.Unmarshal(data, &payload)
	if err != nil {
		return err
	}

	if payload.Key != "username" {
		return ErrWrongPayload
	}

	conn := pool.Get()
	defer conn.Close()

	taken, err := redis.Bool(conn.Do("HEXISTS", "webapp:users", payload.Value))
	if err != nil {
		return ErrDB
	}

	takenTemp, err := redis.Bool(conn.Do("EXISTS", "webapp:temp:users:"+payload.Value))
	if err != nil {
		return ErrDB
	}

	takenPending, err := redis.Bool(conn.Do("EXISTS", "webapp:users:pending:"+payload.Value))
	if err != nil {
		return ErrDB
	}

	payload.Key = "taken"
	if taken || takenTemp || takenPending {
		payload.Value = "true"
	} else {
		payload.Value = "false"
	}

	data, err = json.Marshal(&payload)
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(data)

	return nil
}

func isNameTakenHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	switch r.Method {
	case "POST":
		if err := isNameTaken(w, r); err != nil {
			log.Println(err)
			return
		}
	default:
		http.Error(w, "POST ONLY", http.StatusMethodNotAllowed)
	}
}
