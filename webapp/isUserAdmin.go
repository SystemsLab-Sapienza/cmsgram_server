package main

import (
	"github.com/garyburd/redigo/redis"
)

func isUserAdmin(uid string) (bool, error) {
	conn := pool.Get()
	defer conn.Close()

	user, err := redis.String(conn.Do("HGET", "webapp:users:"+uid, "username"))
	if err != nil {
		return false, ErrDB
	}

	return user == "admin", nil
}
