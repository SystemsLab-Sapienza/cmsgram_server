package main

import (
	"bytes"
	"encoding/json"
	"net/http"
)

// Notifies the bot that the account associated with a user has been deleted
func notifyDelete(uid string) error {
	var (
		payload = struct {
			Key   string
			Value string
		}{"user", uid}
	)

	// Encode data into JSON
	data, err := json.Marshal(&payload)
	if err != nil {
		return err
	}

	// Send the payload to the BotAPI
	_, err = http.Post(config.BotURI+"/account/delete", "application/json", bytes.NewBuffer(data))
	if err != nil {
		return err
	}

	return nil
}
