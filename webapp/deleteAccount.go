package main

import (
	"strings"

	"github.com/garyburd/redigo/redis"
)

func deleteAccount(uid string) error {
	var user = struct {
		Email    string `redis:"email"`
		Username string `redis:"username"`
		Auth     string `redis:"auth"`
	}{}

	conn := pool.Get()
	defer conn.Close()

	data, err := redis.Values(conn.Do("HGETALL", "webapp:users:"+uid))
	if err != nil {
		return ErrDB
	}

	err = redis.ScanStruct(data, &user)
	if err != nil {
		return ErrDB
	}

	lname, err := redis.String(conn.Do("HGET", "webapp:users:data:"+uid, "cognome"))
	if err != nil && err != redis.ErrNil {
		return ErrDB
	}

	// Get the messages sent by the user
	messages, err := redis.Strings(conn.Do("LRANGE", "webapp:users:messages:"+uid, 0, -1))
	if err != nil && err != redis.ErrNil {
		return ErrDB
	}

	// Delete all the data associated with the user
	conn.Send("MULTI")
	conn.Send("HDEL", "webapp:users", user.Username)
	conn.Send("HDEL", "webapp:users:email", user.Email)
	conn.Send("DEL", "webapp:users:auth:session:"+user.Auth)
	conn.Send("DEL", "webapp:users:"+uid)
	conn.Send("DEL", "webapp:users:data:"+uid)
	conn.Send("DEL", "webapp:users:data:email:"+uid)
	conn.Send("DEL", "webapp:users:data:url:"+uid)
	for _, m := range messages {
		conn.Send("DEL", "webapp:messages:"+m)
	}
	conn.Send("DEL", "webapp:users:messages:"+uid)
	if len(lname) != 0 {
		conn.Send("SREM", "webapp:users:info:"+strings.ToLower(lname), uid)
	}
	_, err = conn.Do("EXEC")
	if err != nil {
		return ErrDB
	}

	if err = notifyDelete(uid); err != nil {
		return err
	}

	return nil
}
