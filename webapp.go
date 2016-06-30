package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/garyburd/redigo/redis"
)

// Global variables
var (
	Pool *redis.Pool
)

func init() {
	http.Handle("/assets/", http.FileServer(http.Dir("pages")))
	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/account/delete", accountDeleteHandler)
	http.HandleFunc("/account/password/reset", resetPasswordHandler)
	http.HandleFunc("/account/signin", signinHandler)
	http.HandleFunc("/account/signout", signoutHandler)
	http.HandleFunc("/account/signup", signupHandler)
	http.HandleFunc("/account/verify", accountVerifyHandler)
	http.HandleFunc("/admin/accept", adminAcceptHandler)
	http.HandleFunc("/admin/deny", adminDenyHandler)
	http.HandleFunc("/data/edit", dataEditHandler)
	http.HandleFunc("/data/view", dataViewHandler)
	http.HandleFunc("/isNameTaken", isNameTakenHandler)
	http.HandleFunc("/sendMessage", sendMessageHandler)

	// Create a thread-safe connection pool for redis
	Pool = &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial(Config.RedisDomain, Config.RedisAddress)
			if err != nil {
				return nil, err
			}

			return c, err
		},
	}
}

func main() {
	// TODO get the config file from the command line
	if readConfigFile("/Users/marcofelici/go/src/webapp/webapp.conf") != nil {
		return
	}

	err := os.Chdir(Config.WorkingDirectory)
	if err != nil {
		log.Fatal("main: os.Chdir:", err)
	}

	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("main: http.ListenAndServe:", err)
	}
}
