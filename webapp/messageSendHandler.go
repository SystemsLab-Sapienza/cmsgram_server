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

func messageSend(w http.ResponseWriter, r *http.Request) error {
	var payload = struct {
		Key   string
		Value string
	}{}

	logged, user, err := isLoggedIn(w, r)
	if err != nil {
		return err
	}
	if !logged {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return nil
	}

	msg := r.PostFormValue("message")
	if len(msg) == 0 {
		return ErrFieldEmpty
	}

	// Get the current timestamp
	now := time.Now().Unix()

	conn := pool.Get()
	defer conn.Close()

	// Get the new message ID
	c, err := redis.Int64(conn.Do("INCR", "webapp:messages:counter"))
	if err != nil {
		log.Println("messageSend():", err)
		return ErrDB
	}

	newsid := strconv.FormatInt(c, 10)

	// Save the message on the database
	_, err = conn.Do("HMSET", "webapp:messages:"+newsid, "user_id", user, "timestamp", now, "content", msg)
	if err != nil {
		log.Println("messageSend():", err)
		return ErrDB
	}

	// Push the message ID to the user's message queue
	_, err = conn.Do("LPUSH", "webapp:users:messages:"+user, newsid)
	if err != nil {
		log.Println("messageSend():", err)
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
	_, err = http.Post(config.BotURI+"/message/send", "application/json", bytes.NewReader(data))
	if err != nil {
		return ErrNoServer
	}

	return nil
}

func messageSendHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	switch r.Method {
	case "POST":
		if err := messageSend(w, r); err != nil {
			w.Header().Set("Content-type", "text/plain")
			w.Write([]byte(err.Error()))
			return
		}
		http.Redirect(w, r, "/", http.StatusSeeOther)
	default:
		http.Error(w, "POST ONLY", http.StatusMethodNotAllowed)
	}
}
