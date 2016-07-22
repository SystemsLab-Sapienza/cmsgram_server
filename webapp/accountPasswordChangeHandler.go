package main

import (
	"log"
	"net/http"

	"golang.org/x/crypto/bcrypt"

	"github.com/garyburd/redigo/redis"
)

func accountPasswordChangeGET(w http.ResponseWriter, r *http.Request) error {
	logged, _, err := isLoggedIn(w, r)
	if err != nil {
		return ErrGeneric
	}
	if !logged {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return nil
	}

	templates.ExecuteTemplate(w, "new.html", nil)

	return nil
}

func accountPasswordChangePOST(w http.ResponseWriter, r *http.Request) error {
	var (
		pwd1 = r.PostFormValue("password-old")
		pwd2 = r.PostFormValue("password-new")
		pwd3 = r.PostFormValue("password-new-repeat")
	)

	logged, uid, err := isLoggedIn(w, r)
	if err != nil {
		return ErrGeneric
	}
	if !logged {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return nil
	}

	if pwd1 == "" || pwd2 == "" || pwd3 == "" {
		msg := "Uno o più campi vuoti."
		templates.ExecuteTemplate(w, "new.html", msg)
		return nil
	}

	if pwd2 != pwd3 {
		msg := "I campi della nuova password non coincidono."
		templates.ExecuteTemplate(w, "new.html", msg)
		return nil
	}

	conn := pool.Get()
	defer conn.Close()

	// Get the current hash
	hash, err := redis.Bytes(conn.Do("HGET", "webapp:users:"+uid, "hash"))
	if err != nil && err != redis.ErrNil {
		return ErrDB
	}

	// Check that the supplied password matches the hash
	err = bcrypt.CompareHashAndPassword(hash, []byte(pwd1))
	if err != nil {
		msg := "La password corrente fornita non è valida."
		templates.ExecuteTemplate(w, "new.html", msg)
		return nil
	}

	// Compute the hash for the new password
	pwd2Hash, err := bcrypt.GenerateFromPassword([]byte(pwd2), bcrypt.DefaultCost)
	if err != nil {
		return ErrGeneric
	}

	// Save the new hash
	conn.Do("HSET", "webapp:users:"+uid, "hash", string(pwd2Hash))
	if err != nil && err != redis.ErrNil {
		return ErrDB
	}

	// Invalidate current session
	_, err = conn.Do("HSET", "webapp:users:"+uid, "auth", "")
	if err != nil {
		return ErrDB
	}

	msg := "La password è stata aggiornata."
	templates.ExecuteTemplate(w, "confirm.html", msg)

	return nil
}

func accountPasswordChangeHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	switch r.Method {
	case "GET":
		if err := accountPasswordChangeGET(w, r); err != nil {
			http.Error(w, "Internal error", http.StatusInternalServerError)
			log.Printf("handling %q: %v", r.RequestURI, err)
			return
		}
	case "POST":
		if err := accountPasswordChangePOST(w, r); err != nil {
			http.Error(w, "Internal error", http.StatusInternalServerError)
			log.Printf("handling %q: %v", r.RequestURI, err)
			return
		}
	default:
		http.Error(w, "GET/POST ONLY", http.StatusMethodNotAllowed)
	}
}
