package main

import (
	"flag"
	"log"

	"github.com/garyburd/redigo/redis"
)

var (
	domain     string = "tcp"
	address    string = ":6379"
	email      string
	socketpath string
)

func init() {
	flag.StringVar(&email, "e", "", "Specifies the email of the admin user.")
	flag.StringVar(&socketpath, "s", "", "Connects to redis using the socket provided.")
}

func main() {
	flag.Parse()

	if len(email) == 0 {
		log.Fatal("You need to specify an email address: newadmin -e name@example.com")
	}

	if len(socketpath) != 0 {
		domain = "unix"
		address = socketpath
	}

	c, err := redis.Dial(domain, address)
	if err != nil {
		log.Fatal("redis:", err)
	}
	defer c.Close()

	if err = createAdmin(c, email); err != nil {
		log.Fatal(err)
	}
}
