package main

import (
	"html/template"
	"log"
	"net/http"
	"time"

	// "webapp/auth"

	"github.com/garyburd/redigo/redis"
	"golang.org/x/crypto/bcrypt"
)

func signin(w http.ResponseWriter, r *http.Request) error {
	var (
		user       = r.PostFormValue("username")
		pwd        = r.PostFormValue("password")
		rememberme = r.PostFormValue("rememberme")

		cookie  http.Cookie
		exptime time.Time
	)

	if user == "" {
		return ErrNoUsername
	}
	if pwd == "" {
		return ErrNoPassword
	}

	conn := Pool.Get()
	defer conn.Close()

	// Get the user id, if any
	uid, err := redis.String(conn.Do("HGET", "webapp:users", user))
	if err != nil && err != redis.ErrNil {
		return ErrDB
	} else if uid == "" {
		return ErrWrongCredentials
	}

	hash, err := redis.Bytes(conn.Do("HGET", "webapp:users:"+uid, "hash"))
	if err != nil {
		return ErrDB
	}

	// Wrong password
	err = bcrypt.CompareHashAndPassword(hash, []byte(r.PostFormValue("password")))
	if err != nil {
		return ErrWrongCredentials
	}

	token, err := newSessionToken(32)
	if err != nil {
		return err
	}

	if rememberme == "" {
		exptime = time.Now().AddDate(0, 0, 1)
		cookie = http.Cookie{Name: "auth", Value: token, Path: "/", HttpOnly: true}
	} else {
		exptime = time.Now().AddDate(1, 0, 0)
		cookie = http.Cookie{Name: "auth", Value: token, Path: "/", Expires: exptime, HttpOnly: true}
	}
	http.SetCookie(w, &cookie)

	conn.Send("MULTI")
	conn.Send("SET", "webapp:users:auth:session:"+cookie.Value, uid)
	conn.Send("EXPIREAT", "webapp:users:auth:session:"+cookie.Value, exptime.Unix())
	conn.Send("HSET", "webapp:users:"+uid, "auth", cookie.Value)
	_, err = conn.Do("EXEC")
	if err != nil {
		return ErrDB
	}

	return nil
}

func signinHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	if r.Method == "POST" {
		if err := signin(w, r); err != nil {
			errmsg := err.Error()
			t, err := template.ParseFiles("pages/signin.html")
			if err != nil {
				http.Error(w, "Internal error", http.StatusInternalServerError)
				log.Printf("handling %q: %v", r.RequestURI, err)
				return
			}
			t.Execute(w, errmsg)
			return
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)
	} else {
		http.Error(w, "POST ONLY", http.StatusMethodNotAllowed)
	}
}
