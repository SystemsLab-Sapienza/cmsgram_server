package main

import (
	"log"
	"net/http"
	"regexp"
	"time"

	"github.com/garyburd/redigo/redis"
	"golang.org/x/crypto/bcrypt"
)

func sendResetLink(to string) (token string, err error) {
	// Generate a new reset token
	token, err = newResetToken(32)
	if err != nil {
		return
	}

	subject := "Resetta la tua password"
	body := "Clicka il seguente link per resettare la tua password:\n" +
		config.Domain + "/account/password/reset?token=" + token

	go sendEmail(to, subject, body)

	return
}

func resetPasswordGET(w http.ResponseWriter, r *http.Request) error {
	token := r.FormValue("token")

	// GET request w/o token means the user has requested the form for pwd reset
	if len(token) == 0 {
		http.ServeFile(w, r, "templates/reset.html")
	} else { // GET request with token means the user has clicked the link sent by email
		conn := pool.Get()
		defer conn.Close()

		// Check that token is valid
		_, err := conn.Do("GET", "webapp:auth:reset:"+token)
		if err == redis.ErrNil {
			return ErrBadToken
		} else if err != nil {
			return ErrDB
		}

		templates.ExecuteTemplate(w, "change.html", token)
	}

	return nil
}

func resetPasswordPOST(w http.ResponseWriter, r *http.Request) error {
	const delay = 20
	token := r.PostFormValue("token")

	// Post request w/o token means user has submitted the pwd reset form
	if len(token) == 0 {
		var (
			username = r.FormValue("username")
			email    = r.FormValue("email")
		)

		if len(username) == 0 || len(email) == 0 {
			return ErrFieldEmpty
		}

		// Email address not valid
		ok, err := regexp.Match(`^.+@(.+)?uniroma1.it$`, []byte(email))
		if !ok || err != nil {
			return ErrBadEmail
		}

		conn := pool.Get()
		defer conn.Close()

		// Get the ID associated with the user
		uid, err := redis.String(conn.Do("HGET", "webapp:users", username))
		if err == redis.ErrNil {
			return ErrWrongCredentials
		} else if err != nil {
			return ErrDB
		}

		// Get the email associated with the username
		email2, err := redis.String(conn.Do("HGET", "webapp:users:"+uid, "email"))
		if err != nil {
			return ErrDB
		}

		// Email and username must match
		if email != email2 {
			return ErrNoMatch
		}

		// Send reset link to the email address provided
		token, err := sendResetLink(email)
		if err != nil {
			return err
		}
		exptime := time.Now().Add(time.Minute * delay).Unix()

		conn.Send("MULTI")
		conn.Send("SET", "webapp:auth:reset:"+token, uid)
		conn.Send("EXPIREAT", "webapp:auth:reset"+token, exptime)
		_, err = conn.Do("EXEC")
		if err != nil {
			return ErrDB
		}

		msg := "Il link di reset è stato inviato all'indirizzo email fornito."
		templates.ExecuteTemplate(w, "confirm.html", msg)
	} else { // Post request w/ token means the user has submitted the change password form
		var (
			pwd1 = r.PostFormValue("password1")
			pwd2 = r.PostFormValue("password2")
		)

		conn := pool.Get()
		defer conn.Close()

		// Check that token is valid
		uid, err := redis.String(conn.Do("GET", "webapp:auth:reset:"+token))
		if err == redis.ErrNil {
			return ErrBadToken
		} else if err != nil {
			return ErrDB
		}

		if len(pwd1) == 0 || len(pwd2) == 0 {
			return ErrFieldEmpty
		}

		// Password must match
		if pwd1 != pwd2 {
			return ErrNoMatch
		}

		// Hash the new password
		hash, err := bcrypt.GenerateFromPassword([]byte(pwd1), bcrypt.DefaultCost)
		if err != nil {
			return ErrGeneric
		}

		// Save it on the DB
		_, err = conn.Do("HSET", "webapp:users:"+uid, "hash", hash)
		if err != nil {
			return ErrDB
		}

		// Delete token
		_, err = conn.Do("DEL", "webapp:auth:reset:"+token)
		if err != nil {
			return ErrDB
		}

		// Invalidate current session
		_, err = conn.Do("HSET", "webapp:users:"+uid, "auth", "")
		if err != nil {
			return ErrDB
		}

		msg := "La password è stata modificata."
		templates.ExecuteTemplate(w, "confirm.html", msg)
	}

	return nil
}

func resetPasswordHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	switch r.Method {
	case "GET":
		if err := resetPasswordGET(w, r); err != nil {
			log.Println(err)
			return
		}
	case "POST":
		if err := resetPasswordPOST(w, r); err != nil {
			log.Println(err)
			return
		}
	default:
		http.Error(w, "POST ONLY", http.StatusMethodNotAllowed)
	}
}
