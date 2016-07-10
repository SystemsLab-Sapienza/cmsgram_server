package main

import (
	"net/http"

	"github.com/garyburd/redigo/redis"
)

func renderUserData(w http.ResponseWriter, r *http.Request, tname string) error {
	var user = struct {
		Email      string   `redis:"email"`
		EmailAltre []string `redis:"-"`
		Nome       string   `redis:"nome"`
		Cognome    string   `redis:"cognome"`
		Indirizzo  string   `redis:"indirizzo"`
		Telefono   string   `redis:"telefono"`
		Tipo       string   `redis:"tipo"`
		Sito       string   `redis:"sito"`
		SitoAltri  []string `redis:"-"`
	}{}

	logged, uid, err := isLoggedIn(w, r)
	if err != nil {
		return ErrGeneric
	}
	if !logged {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return nil
	}

	conn := Pool.Get()
	defer conn.Close()

	user.Email, err = redis.String(conn.Do("HGET", "webapp:users:"+uid, "email"))
	if err != nil {
		return ErrDB
	}

	data, err := redis.Values(conn.Do("HGETALL", "webapp:users:data:"+uid))
	if err != nil {
		return ErrDB
	}

	err = redis.ScanStruct(data, &user)
	if err != nil {
		return ErrDB
	}

	user.EmailAltre, err = redis.Strings(conn.Do("LRANGE", "webapp:users:data:email:"+uid, 0, -1))
	if err != nil {
		return ErrDB
	}

	user.SitoAltri, err = redis.Strings(conn.Do("LRANGE", "webapp:users:data:url:"+uid, 0, -1))
	if err != nil {
		return ErrDB
	}

	if err = templates.ExecuteTemplate(w, tname, user); err != nil {
		return ErrGeneric
	}

	return nil
}
