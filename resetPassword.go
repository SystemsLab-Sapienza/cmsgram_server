package main

import (
	"html/template"
	"log"
	"net/http"
	"net/smtp"
	"time"

	"github.com/garyburd/redigo/redis"
	"golang.org/x/crypto/bcrypt"
)

func sendResetLink(email string) (string, error) {
	token, err := newResetToken(32)
	if err != nil {
		return "", err
	}
	text := "Clicka il link seguente per resettare la tua password.\n" + "http://localhost:8080/auth/reset?token=" + token

	smtpAuth := smtp.PlainAuth("", Config.EmailUsername, Config.EmailPassword, "smtp.gmail.com")

	to := []string{email}

	err = smtp.SendMail(Config.EmailServer, smtpAuth, "", to, []byte(text))
	if err != nil {
		return "", err
	}

	return token, nil
}

func resetPassword(w http.ResponseWriter, r *http.Request) error {
	if r.Method == "GET" {
		token := r.FormValue("token")

		// GET request w/o token means the user has requested the form for pwd reset
		if token == "" {
			http.ServeFile(w, r, "pages/reset.html")
		} else { // GET request with token means the user has clicked the link sent by email
			conn := Pool.Get()
			defer conn.Close()

			// Check that token is valid
			_, err := conn.Do("GET", "webapp:auth:reset:"+token)
			if err == redis.ErrNil {
				return ErrBadToken
			} else if err != nil {
				return ErrDB
			}

			t, err := template.ParseFiles("pages/change.html")
			if err != nil {
				http.Error(w, "Internal error", http.StatusInternalServerError)
				log.Printf("handling %q: %v", r.RequestURI, err)
				return err
			}
			t.Execute(w, token)
		}
	} else if r.Method == "POST" {
		const delay = 20
		token := r.PostFormValue("token")

		// Post request w/o token means user has submitted the pwd reset form
		if token == "" {
			var (
				username = r.FormValue("username")
				email    = r.FormValue("email")
			)

			if username == "" || email == "" {
				return ErrFieldEmpty
			}

			conn := Pool.Get()
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

			// TODO use more specific error
			// Send reset link to the email address provided
			token, err := sendResetLink(email)
			if err != nil {
				return ErrGeneric
			}
			exptime := time.Now().Add(time.Minute * delay).Unix()

			conn.Send("MULTI")
			conn.Send("SET", "webapp:auth:reset:"+token, uid)
			conn.Send("EXPIREAT", "webapp:auth:reset"+token, exptime)
			_, err = conn.Do("EXEC")
			if err != nil {
				return ErrDB
			}

			// TODO return page saying a reset link has been sent
			http.Redirect(w, r, "/", http.StatusSeeOther)
		} else { // Post request w/ token means the user has submitted the change password form
			var (
				pwd1 = r.PostFormValue("password1")
				pwd2 = r.PostFormValue("password2")
			)

			conn := Pool.Get()
			defer conn.Close()

			// Check that token is valid
			uid, err := redis.String(conn.Do("GET", "webapp:auth:reset:"+token))
			if err == redis.ErrNil {
				return ErrBadToken
			} else if err != nil {
				return ErrDB
			}

			if pwd1 == "" || pwd2 == "" {
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
			// TODO return page saying the password has been reset
			http.Redirect(w, r, "/", http.StatusSeeOther)
		}
	} else {
		http.Error(w, "GET/POST ONLY", http.StatusMethodNotAllowed)
	}

	return nil
}

func resetPasswordHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	if err := resetPassword(w, r); err != nil {
		errmsg := err.Error()
		log.Println(errmsg)
		return
	}
}
