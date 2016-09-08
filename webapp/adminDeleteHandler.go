package main

import (
	"log"
	"net/http"
)

func adminDelete(w http.ResponseWriter, r *http.Request) error {
	var ID = r.PostFormValue("user_id")

	logged, uid, err := isLoggedIn(w, r)
	if err != nil {
		return ErrGeneric
	}
	if !logged {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return nil
	}

	admin, err := isUserAdmin(uid)
	if err != nil {
		return err
	} else if !admin {
		http.NotFound(w, r)
		return nil
	}

	if err := deleteAccount(ID); err != nil {
		return err
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
	return nil
}

func adminDeleteHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	switch r.Method {
	case "POST":
		if err := adminDelete(w, r); err != nil {
			log.Println(err)
			return
		}
	default:
		http.Error(w, "POST ONLY", http.StatusMethodNotAllowed)
	}
}
