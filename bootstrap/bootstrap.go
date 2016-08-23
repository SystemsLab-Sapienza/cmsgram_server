package main

import (
	"encoding/json"
	"io/ioutil"
	"strconv"
	"strings"

	"github.com/garyburd/redigo/redis"
)

func bootstrap(conn redis.Conn, filepath string) error {
	var docenti = []struct {
		Tipo     string `redis:"tipo"`
		Nome     string `redis:"nome"`
		Cognome  string `redis:"cognome"`
		Email    string `redis:"email"`
		Telefono string `redis:"telefono"`
		Url      string `redis:"url"`
	}{}

	data, err := ioutil.ReadFile(filepath)
	if err != nil {
		return err
	}

	err = json.Unmarshal(data, &docenti)
	if err != nil {
		return err
	}

	// Delete previous keys, if any
	err = deleteKeys(conn, "webapp:docenti*")
	if err != nil {
		return err
	}

	_, err = conn.Do("SET", "webapp:docenti:counter", 1000)
	if err != nil {
		return err
	}

	for _, d := range docenti {
		lname := strings.ToLower(d.Cognome)

		uid, err := redis.Int(conn.Do("INCR", "webapp:docenti:counter"))
		if err != nil {
			return err
		}

		conn.Send("MULTI")
		conn.Send("SADD", "webapp:docenti", uid)
		conn.Send("SADD", "webapp:docenti:"+lname, uid)
		conn.Send("HMSET", redis.Args{}.Add("webapp:docenti:"+strconv.Itoa(uid)).AddFlat(&d)...)
		_, err = conn.Do("EXEC")
		if err != nil {
			return err
		}
	}

	return nil
}
