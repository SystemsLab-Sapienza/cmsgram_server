package main

import (
	"log"
	"net/http"
)

func accountDelete(w http.ResponseWriter, r *http.Request) error {
	logged, uid, err := isLoggedIn(w, r)
	if err != nil {
		return ErrGeneric
	}
	if !logged {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return nil
	}

	if err := deleteAccount(uid); err != nil {
		return err
	}

	msg := "L'account è stato eliminato."
	templates.ExecuteTemplate(w, "confirm.html", msg)

	return nil
}

func accountDeleteHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	switch r.Method {
	case "POST":
		if err := accountDelete(w, r); err != nil {
			http.Error(w, "Internal error", http.StatusInternalServerError)
			log.Printf("handling %q: %v", r.RequestURI, err)
			return
		}
	default:
		http.Error(w, "POST ONLY", http.StatusMethodNotAllowed)
	}
}
