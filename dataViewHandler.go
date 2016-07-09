package main

import (
	"log"
	"net/http"
)

func dataView(w http.ResponseWriter, r *http.Request) error {
	return renderUserData(w, r, "view.html")
}

func dataViewHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	switch r.Method {
	case "GET":
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")
		w.Header().Set("Content-Type", "text/html")

		if err := dataView(w, r); err != nil {
			http.Error(w, "Internal error", http.StatusInternalServerError)
			log.Printf("handling %q: %v", r.RequestURI, err)
			return
		}
	default:
		http.Error(w, "GET ONLY", http.StatusMethodNotAllowed)
	}
}
