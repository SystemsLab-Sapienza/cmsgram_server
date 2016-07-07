package main

import (
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/garyburd/redigo/redis"
)

// Global variables
var (
	Pool      *redis.Pool
	templates *template.Template
)

func init() {
	http.Handle("/assets/", http.FileServer(http.Dir("templates")))
	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/account/delete", accountDeleteHandler)
	http.HandleFunc("/account/email/verify", accountVerifyHandler)
	http.HandleFunc("/account/password/reset", resetPasswordHandler)
	http.HandleFunc("/account/signin", signinHandler)
	http.HandleFunc("/account/signout", signoutHandler)
	http.HandleFunc("/account/signup", signupHandler)
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

	// Define templates functions
	funcMap := template.FuncMap{
		"isLast": func(i, n int) bool {
			if i == n-1 {
				return true
			}
			return false
		},
		"strcmp": func(a, b string) bool {
			return a == b
		},
	}

	// Parse templates
	templates = template.Must(template.New("templates").Funcs(funcMap).ParseFiles(
		"templates/admin.html",
		"templates/change.html",
		"templates/confirm.html",
		"templates/edit.html",
		"templates/email.tpl",
		"templates/home.html",
		// "templates/reset.html",
		"templates/signup.html",
		"templates/signin.html",
		"templates/view.html"))

	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("main: http.ListenAndServe:", err)
	}
}
