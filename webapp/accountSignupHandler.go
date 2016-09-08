package main

import (
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/garyburd/redigo/redis"
	"golang.org/x/crypto/bcrypt"
)

// Sends an authentication link to the given email and returns the token
func sendAuthLink(to string) (token string, err error) {
	// Generate a new verification token
	token, err = newActivationToken(32)
	if err != nil {
		return
	}

	subject := "Link di verifica"
	body := "Clicka il seguente link per verificare la tua email:\n" +
		config.Domain + "/account/email/verify?token=" + token

	go sendEmail(to, subject, body)

	return
}

func signup(w http.ResponseWriter, r *http.Request) error {
	var (
		user  = strings.ToLower(r.PostFormValue("username"))
		pwd1  = r.PostFormValue("password1")
		pwd2  = r.PostFormValue("password2")
		email = strings.ToLower(r.PostFormValue("email"))
	)

	// One or more fields empty
	if len(user) == 0 || len(pwd1) == 0 || len(pwd2) == 0 || len(email) == 0 {
		return ErrFieldEmpty
	}

	// Password fields must match
	if pwd1 != pwd2 {
		return ErrNoMatch
	}

	// Email address not valid
	ok, err := regexp.Match(`^.+@(.+)?uniroma1.it$`, []byte(email))
	if !ok || err != nil {
		return ErrBadEmail
	}

	conn := pool.Get()
	defer conn.Close()

	// Email addressed already in use
	taken, err := redis.Bool(conn.Do("HEXISTS", "webapp:users:email", email))
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

	token, err := sendAuthLink(email)
	if err != nil {
		return err
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

func signupHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	switch r.Method {
	case "POST":
		if err := signup(w, r); err != nil {
			errmsg := err.Error()
			templates.ExecuteTemplate(w, "signup.html", errmsg)
			return
		}

		msg := "E' stato inviato un link per verificare l'indirizzo email fornito."
		templates.ExecuteTemplate(w, "confirm.html", msg)
	case "GET":
		templates.ExecuteTemplate(w, "signup.html", nil)
	default:
		http.Error(w, "GET/POST ONLY", http.StatusMethodNotAllowed)
	}
}
