package main

import "github.com/garyburd/redigo/redis"

type account struct {
	ID       string
	Username string
	Fullname string
}

func getFullName(uid string) (string, error) {
	var name, lname string

	conn := pool.Get()
	defer conn.Close()

	// Get first and last name
	values, err := redis.Values(conn.Do("HMGET", "webapp:users:data:"+uid, "nome", "cognome"))
	if err != nil {
		return "", err
	}

	_, err = redis.Scan(values, &name, &lname)
	if err != nil {
		return "", err
	}

	return name + " " + lname, nil
}

func getAccounts() ([]account, error) {
	var accounts = make([]account, 0)

	conn := pool.Get()
	defer conn.Close()

	values, err := redis.Strings(conn.Do("HGETALL", "webapp:users"))
	if err != nil {
		return nil, err
	}

	for i := 0; i < len(values); i += 2 {
		username := values[i]
		ID := values[i+1]

		// Skip admin account
		if username == "admin" {
			continue
		}

		name, err := getFullName(ID)
		if err != nil {
			return nil, err
		}

		accounts = append(accounts, account{ID, username, name})
	}

	return accounts, nil
}
