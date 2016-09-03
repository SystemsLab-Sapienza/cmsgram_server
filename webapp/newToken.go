package main

import (
	"bitbucket.org/ansijax/rfidlab_telegramdi_backend/auth"

	"github.com/garyburd/redigo/redis"
)

// Generates a new length-character string in base 36 that's checked to be unique in the DB
func newUniqueToken(length int, base string) (string, error) {
	conn := pool.Get()
	defer conn.Close()

	token := auth.NewBase36(length)
	res, err := redis.String(conn.Do("GET", base+token))
	if err != nil && err != redis.ErrNil {
		return "", ErrDB
	}

	for i := 0; i < 1e3; i++ {
		token = auth.NewBase36(length)
		res, err = redis.String(conn.Do("GET", base+token))
		if err != nil && err != redis.ErrNil {
			return "", ErrDB
		}

		if len(res) == 0 {
			return token, nil
		}
	}

	return "", ErrGeneric
}

func newActivationToken(len int) (string, error) {
	return newUniqueToken(len, "webapp:temp:token:")
}

func newLoginToken(len int) (string, error) {
	return newUniqueToken(len, "webapp:auth:login:")
}

func newResetToken(len int) (string, error) {
	return newUniqueToken(len, "webapp:auth:reset:")
}

func newSessionToken(len int) (string, error) {
	return newUniqueToken(len, "webapp:auth:session:")
}
