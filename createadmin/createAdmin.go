package main

import (
	"errors"
	"regexp"

	"github.com/garyburd/redigo/redis"
	"golang.org/x/crypto/bcrypt"
)

func createAdmin(conn redis.Conn, email string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	ok, err := regexp.Match(`^.+@.+\..{2,}$`, []byte(email))
	if !ok || err != nil {
		return errors.New("Email address not valid.")
	}

	conn.Send("MULTI")
	conn.Send("SET", "webapp:users:counter", 1000)
	conn.Send("HSET", "webapp:users", "admin", 1000)
	conn.Send("HMSET", "webapp:users:1000", "email", email, "username", "admin", "hash", hash)
	_, err = conn.Do("EXEC")
	if err != nil {
		return err
	}

	return nil
}
