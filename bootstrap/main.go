package main

import (
	"flag"
	"log"

	"github.com/garyburd/redigo/redis"
)

var (
	domain     string = "tcp"
	address    string = ":6379"
	filepath   string
	socketpath string
)

func init() {
	flag.StringVar(&filepath, "i", "", "Specifies the path to the file to be imported.")
	flag.StringVar(&socketpath, "s", "", "Connects to redis using the socket provided.")
}

func main() {
	flag.Parse()

	if filepath == "" {
		log.Fatal("You need to specify a file to import: bootstrap -i /path/to/file")
	}

	if socketpath != "" {
		domain = "unix"
		address = socketpath
	}

	c, err := redis.Dial(domain, address)
	if err != nil {
		log.Fatal("redis:", err)
	}
	defer c.Close()

	err = bootstrap(c, filepath)
	if err != nil {
		log.Fatal(err)
	}
}
