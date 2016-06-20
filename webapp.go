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
	http.HandleFunc("/account/accept", accountAcceptHandler)
	http.HandleFunc("/account/activate", accountActivateHandler)
	http.HandleFunc("/account/delete", accountDeleteHandler)
	http.HandleFunc("/account/deny", accountDenyHandler)
	http.HandleFunc("/auth/reset", resetPasswordHandler)
	http.HandleFunc("/data/view", dataViewHandler)
	http.HandleFunc("/data/edit", dataEditHandler)
	http.HandleFunc("/isNameTaken", isNameTakenHandler)
	http.HandleFunc("/sendMessage", sendMessageHandler)
	http.HandleFunc("/signin", signinHandler)
	http.HandleFunc("/signout", signoutHandler)
	http.HandleFunc("/signup", signupHandler)

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
		return
	}

	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("main: http.ListenAndServe:", err)
		return
	}
}
