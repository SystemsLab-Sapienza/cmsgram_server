package main

import (
	"bytes"
	"encoding/json"
	"net/http"
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

		conn := Pool.Get()
		defer conn.Close()

		// Push the message to the user's message queue
		_, err = conn.Do("LPUSH", "webapp:messages:"+user, msg)
		if err != nil {
			return ErrDB
		}

		// Create the JSON payload
		payload.Key = "message"
		payload.Value = msg
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
