package main

import (
	"log"
	"net/http"
	"strings"

	"github.com/garyburd/redigo/redis"
)

func dataEditGet(w http.ResponseWriter, r *http.Request) error {
	return renderUserData(w, r, "edit.html")
}

func dataEditPost(w http.ResponseWriter, r *http.Request) error {
	var user = struct {
		Email      string   `redis:"email"`
		EmailAltre []string `redis:"-"`
		Nome       string   `redis:"nome"`
		Cognome    string   `redis:"cognome"`
		Tipo       string   `redis:"tipo"`
		Indirizzo  string   `redis:"indirizzo"`
		Telefono   string   `redis:"telefono"`
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

	if err := r.ParseForm(); err != nil {
		return ErrGeneric
	}

	for k, v := range r.Form {
		switch k {
		case "email":
			user.Email = v[0]
			for _, e := range v[1:] {
				if e != "" {
					user.EmailAltre = append(user.EmailAltre, e)
				}
			}
			break
		case "fname":
			user.Nome = v[0]
			break
		case "lname":
			user.Cognome = v[0]
			break
		case "type":
			user.Tipo = v[0]
			break
		case "address":
			user.Indirizzo = v[0]
			break
		case "phone":
			user.Telefono = v[0]
			break
		case "url":
			user.Sito = v[0]
			for _, s := range v[1:] {
				if s != "" {
					user.SitoAltri = append(user.SitoAltri, s)
				}
			}
			break
		default:
			log.Println("postData: Form key not valid:", k)
			break
		}
	}

	conn := Pool.Get()
	defer conn.Close()

	// Last name changed
	lname, err := redis.String(conn.Do("HGET", "webapp:users:data:"+uid, "cognome"))
	if lname != "" && lname != user.Cognome {
		// TODO usa un set per gestire le collisioni, gestisci errori
		_, err = conn.Do("HSET", "webapp:info", strings.ToLower(user.Cognome), uid)
		if err != nil {
			return ErrDB
		}
	}

	_, err = conn.Do("HMSET", redis.Args{}.Add("webapp:users:data:"+uid).AddFlat(&user)...)
	if err != nil {
		return ErrDB
	}

	_, err = conn.Do("DEL", "webapp:users:data:email:"+uid)
	if err != nil {
		return ErrDB
	}
	if len(user.EmailAltre) > 0 {
		_, err = conn.Do("RPUSH", redis.Args{}.Add("webapp:users:data:email:"+uid).AddFlat(user.EmailAltre)...)
		if err != nil {
			return ErrDB
		}
	}

	_, err = conn.Do("DEL", "webapp:users:data:url:"+uid)
	if err != nil {
		return ErrDB
	}
	if len(user.SitoAltri) > 0 {
		_, err = conn.Do("RPUSH", redis.Args{}.Add("webapp:users:data:url:"+uid).AddFlat(user.SitoAltri)...)
		if err != nil {
			return ErrDB
		}
	}

	http.Redirect(w, r, "/data/view", http.StatusSeeOther)
	return nil
}

func dataEditHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	if r.Method == "GET" {
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")
		w.Header().Set("Content-Type", "text/html")

		dataEditGet(w, r)
	} else if r.Method == "POST" {
		if err := dataEditPost(w, r); err != nil {
			http.Error(w, "Internal error", http.StatusInternalServerError)
			log.Printf("handling %q: %v", r.RequestURI, err)
			return
		}
	} else {
		http.Error(w, "GET/POST ONLY", http.StatusMethodNotAllowed)
	}
}
