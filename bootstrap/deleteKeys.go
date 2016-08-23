package main

import "github.com/garyburd/redigo/redis"

// Scans redis for keys matching the pattern and deletes matches in bulk
func deleteKeys(conn redis.Conn, pattern string) (err error) {
	var (
		count   int64
		keys    []string
		matches []string
	)

	data, err := redis.Values(conn.Do("SCAN", 0, "MATCH", pattern))
	if err != nil {
		return
	}

	data, err = redis.Scan(data, &count, &keys)
	if err != nil {
		return
	}

	matches = append(matches, keys...)

	for count != 0 {
		data, err = redis.Values(conn.Do("SCAN", count, "MATCH", pattern))
		if err != nil {
			return
		}

		data, err = redis.Scan(data, &count, &keys)
		if err != nil {
			return
		}

		matches = append(matches, keys...)
	}

	if len(matches) > 0 {
		_, err = conn.Do("DEL", redis.Args{}.AddFlat(matches)...)
		if err != nil {
			return err
		}
	}

	return
}
