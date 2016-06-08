package main

import (
	"bytes"
	"net/http"
)

func sendMessage(w http.ResponseWriter, r *http.Request) error {
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

		// Push the message on the message queue
		_, err = conn.Do("LPUSH", "webapp:messages:"+user, msg)
		if err != nil {
			return ErrDB
		}

		// TODO use more specific error
		_, err := http.Post(Config.SendMessageEndpoint, "text/plain", bytes.NewReader([]byte(msg)))
		if err != nil {
			return err
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
