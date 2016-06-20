package main

import (
	"webapp/auth"

	"github.com/garyburd/redigo/redis"
)

// Generates a new len-character string in base 36 that's checked to be unique in the DB
func newUniqueToken(len int, base string) (string, error) {
	conn := Pool.Get()
	defer conn.Close()

	token := auth.NewBase36(len)
	res, err := redis.String(conn.Do("GET", base+token))
	if err != nil && err != redis.ErrNil {
		return "", ErrDB
	}
	for res != "" {
		token = auth.NewBase36(len)
		res, err = redis.String(conn.Do("GET", base+token))
		if err != nil && err != redis.ErrNil {
			return "", ErrDB
		}
	}

	return token, nil
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
