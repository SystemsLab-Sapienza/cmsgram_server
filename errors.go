package main

import (
	"errors"
)

var (
	ErrBadToken         = errors.New("Token scaduto o non valido.")
	ErrBadEmail         = errors.New("Indirizzo email non valido.")
	ErrDB               = errors.New("Errore database. Riprova più tardi.")
	ErrEMailTaken       = errors.New("Indirizzo email già in uso.")
	ErrFieldEmpty       = errors.New("Uno o più campi vuoti.")
	ErrGeneric          = errors.New("Errore interno.")
	ErrNoPassword       = errors.New("Il campo password non può essere vuoto.")
	ErrNoUsername       = errors.New("Il campo username non può essere vuoto.")
	ErrNameTaken        = errors.New("Nome utente già in uso.")
	ErrWrongCredentials = errors.New("Credenziali non valide.")
)
