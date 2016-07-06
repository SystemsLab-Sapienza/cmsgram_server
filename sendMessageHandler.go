package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/garyburd/redigo/redis"
)

func sendMessage(w http.ResponseWriter, r *http.Request) error {
	var payload = struct {
		Key   string
		Value string
	}{}

	logged, user, err := isLoggedIn(w, r)
	if err != nil {
		return err
	}
	if logged {
		msg := r.PostFormValue("message")
		if msg == "" {
			return ErrFieldEmpty
		}

		// Get the current timestamp
		now := time.Now().Unix()

		conn := Pool.Get()
		defer conn.Close()

		// Get the new message id
		c, err := redis.Int64(conn.Do("INCR", "webapp:messages:counter"))
		if err != nil {
			log.Println("sendMessage():", err)
			return ErrDB
		}

		newsid := strconv.FormatInt(c, 10)

		// Save the message on the database
		_, err = conn.Do("HMSET", "webapp:messages:"+newsid, "user_id", user, "timestamp", now, "content", msg)
		if err != nil {
			log.Println("sendMessage():", err)
			return ErrDB
		}

		// Push the message id to the user's message queue
		_, err = conn.Do("LPUSH", "webapp:users:messages:"+user, newsid)
		if err != nil {
			log.Println("sendMessage():", err)
			return ErrDB
		}

		// Create the JSON payload
		payload.Key = "message"
		payload.Value = newsid
		data, err := json.Marshal(&payload)
		if err != nil {
			return ErrGeneric
		}

		// Send the payload
		_, err = http.Post(Config.SendMessageEndpoint, "application/json", bytes.NewReader(data))
		if err != nil {
			return ErrNoServer
		}
	}

	return nil
}

func sendMessageHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	if r.Method == "POST" {
		if err := sendMessage(w, r); err != nil {
			w.Header().Set("Content-type", "text/plain")
			w.Write([]byte(err.Error()))
			return
		}
		http.Redirect(w, r, "/", http.StatusSeeOther)
	} else {
		http.Error(w, "POST ONLY", http.StatusMethodNotAllowed)
	}
}
