package main

import (
	"net/http"
	"time"

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

	if len(user) == 0 {
		return ErrNoUsername
	}
	if len(pwd) == 0 {
		return ErrNoPassword
	}

	conn := pool.Get()
	defer conn.Close()

	// Get the user ID, if any
	uid, err := redis.String(conn.Do("HGET", "webapp:users", user))
	if err != nil && err != redis.ErrNil {
		return ErrDB
	}

	// Check if username is actually an email address
	if len(uid) == 0 {
		uid, err = redis.String(conn.Do("HGET", "webapp:users:email", user))
		if err != nil && err != redis.ErrNil {
			return ErrDB
		}
	}

	if len(uid) == 0 {
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

	if len(rememberme) == 0 {
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

	switch r.Method {
	case "GET":
		templates.ExecuteTemplate(w, "signin.html", nil)
	case "POST":
		if err := signin(w, r); err != nil {
			templates.ExecuteTemplate(w, "signin.html", err)
			return
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)
	default:
		http.Error(w, "POST ONLY", http.StatusMethodNotAllowed)
	}
}
