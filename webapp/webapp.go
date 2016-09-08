package main

import (
	"flag"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/garyburd/redigo/redis"
)

// Global variables
var (
	pool      *redis.Pool
	templates *template.Template

	flagConfigFile string
)

func init() {
	flag.StringVar(&flagConfigFile, "c", "", "Specifies the path to the config file.")

	http.Handle("/assets/", http.FileServer(http.Dir("templates")))
	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/account/delete", accountDeleteHandler)
	http.HandleFunc("/account/email/verify", accountVerifyHandler)
	http.HandleFunc("/account/password/change", accountPasswordChangeHandler)
	http.HandleFunc("/account/password/reset", resetPasswordHandler)
	http.HandleFunc("/account/signin", signinHandler)
	http.HandleFunc("/account/signout", signoutHandler)
	http.HandleFunc("/account/signup", signupHandler)
	http.HandleFunc("/admin/accept", adminAcceptHandler)
	http.HandleFunc("/admin/delete", adminDeleteHandler)
	http.HandleFunc("/admin/deny", adminDenyHandler)
	http.HandleFunc("/data/edit", dataEditHandler)
	http.HandleFunc("/data/view", dataViewHandler)
	http.HandleFunc("/isNameTaken", isNameTakenHandler)
	http.HandleFunc("/message/send", messageSendHandler)

	// Create a thread-safe connection pool for redis
	pool = &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial(config.RedisDomain, config.RedisAddress)
			if err != nil {
				return nil, err
			}

			return c, err
		},
	}
}

func main() {
	flag.Parse()

	if len(flagConfigFile) == 0 {
		log.Fatal("You need to specify a configuration file: webapp -c /path/to/file")
	}

	if err := readConfigFile(flagConfigFile); err != nil {
		log.Fatal("Error while reading config file:", err)
	}

	// Change the working directory
	if err := os.Chdir(config.WorkingDirectory); err != nil {
		log.Fatal(err)
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

	// Parse the templates
	templates = template.Must(template.New("templates").Funcs(funcMap).ParseFiles(
		"templates/admin.html",
		"templates/change.html",
		"templates/confirm.html",
		"templates/edit.html",
		"templates/email.tpl",
		"templates/home.html",
		"templates/new.html",
		// "templates/reset.html",
		"templates/signup.html",
		"templates/signin.html",
		"templates/view.html"))

	log.Fatal("main: http.ListenAndServe:", http.ListenAndServe(":8080", nil))
}
