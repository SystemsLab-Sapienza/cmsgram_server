package main

import (
	"flag"
	"fmt"

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
		fmt.Println("You need to specify a file to import: bootstrap -i /path/to/file")
		return
	}

	if socketpath != "" {
		domain = "unix"
		address = socketpath
	}

	c, err := redis.Dial(domain, address)
	if err != nil {
		fmt.Println("redis:", err)
		return
	}
	defer c.Close()

	err = bootstrap(c, filepath)
	if err != nil {
		fmt.Println(err)
	}
}
