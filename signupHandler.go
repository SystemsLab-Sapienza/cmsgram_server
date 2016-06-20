package main

import (
	"html/template"
	"log"
	"net/http"
	"net/smtp"
	"regexp"
	"time"

	"github.com/garyburd/redigo/redis"
	"golang.org/x/crypto/bcrypt"
)

// TODO the text and the URL must be taken from the config file.
// Send an authentication link to the given email and returns the token
func sendAuthLink(email string) (string, error) {
	authString, err := newActivationToken(32)
	if err != nil {
		return "", err
	}
	text := "Clicka il link seguente per verificare la tua email.\n" + "http://localhost:8080/account/activate?token=" + authString

	smtpAuth := smtp.PlainAuth("", Config.EmailUsername, Config.EmailPassword, "smtp.gmail.com")

	to := []string{email}

	err = smtp.SendMail(Config.EmailServer, smtpAuth, "", to, []byte(text))
	if err != nil {
		return "", err
	}

	return authString, nil
}

func signup(w http.ResponseWriter, r *http.Request) error {
	var (
		user  = r.PostFormValue("username")
		pwd1  = r.PostFormValue("password1")
		pwd2  = r.PostFormValue("password2")
		email = r.PostFormValue("email")
	)

	// One or more fields empty
	if user == "" || pwd1 == "" || pwd2 == "" || email == "" {
		return ErrFieldEmpty
	}

	// Password fields must match
	if pwd1 != pwd2 {
		return ErrNoMatch
	}

	// TODO get pattern from config file
	// Email address not valid
	ok, err := regexp.Match(`^.+@(.+)?uniroma1.it$`, []byte(email))
	if !ok || err != nil {
		return ErrBadEmail
	}

	conn := Pool.Get()
	defer conn.Close()

	// Email addressed already in use
	taken, err := redis.Bool(conn.Do("SISMEMBER", "webapp:users:email", email))
	if err != nil {
		return ErrDB
	}
	if taken {
		return ErrEMailTaken
	}

	// Email addressed already in the signup process
	taken, err = redis.Bool(conn.Do("EXISTS", "webapp:temp:email:"+email))
	if err != nil {
		return ErrDB
	}
	if taken {
		return ErrEMailTaken
	}

	// Username already taken
	taken, err = redis.Bool(conn.Do("HEXISTS", "webapp:users", user))
	if err != nil {
		return ErrDB
	}
	if taken {
		return ErrNameTaken
	}

	// Username already in the signup process
	taken, err = redis.Bool(conn.Do("EXISTS", "webapp:temp:users:"+user))
	if err != nil {
		return ErrDB
	}
	if taken {
		return ErrNameTaken
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(pwd1), bcrypt.DefaultCost)
	if err != nil {
		return ErrGeneric
	}

	if Config.EmailTestAddress != "" {
		email = Config.EmailTestAddress
	}
	// TODO Return a more specific error
	token, err := sendAuthLink(email)
	if err != nil {
		return ErrGeneric
	}
	exptime := time.Now().Add(time.Hour).Unix()

	// Store the data temporarily, pending validation
	conn.Send("MULTI")
	conn.Send("HMSET", "webapp:temp:token:"+token, "email", email, "username", user, "hash", hash)
	conn.Send("EXPIREAT", "webapp:temp:token:"+token, exptime)
	conn.Send("SET", "webapp:temp:email:"+email, user)
	conn.Send("EXPIREAT", "webapp:temp:email:"+email, exptime)
	conn.Send("SET", "webapp:temp:users:"+user, email)
	conn.Send("EXPIREAT", "webapp:temp:users:"+user, exptime)
	_, err = conn.Do("EXEC")
	if err != nil {
		return ErrDB
	}

	return nil
}

// TODO Use a double handler for better error handling
func signupHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	if r.Method == "POST" {
		if err := signup(w, r); err != nil {
			errmsg := err.Error()
			t, err := template.ParseFiles("pages/signup.html")
			if err != nil {
				http.Error(w, "Internal error", http.StatusInternalServerError)
				log.Printf("handling %q: %v", r.RequestURI, err)
				return
			}
			t.Execute(w, errmsg)
			return
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)
	} else if r.Method == "GET" {
		t, err := template.ParseFiles("pages/signup.html")
		if err != nil {
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}
		t.Execute(w, nil)
	} else {
		http.Error(w, "GET/POST ONLY", http.StatusMethodNotAllowed)
	}
}
