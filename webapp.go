package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"time"

	"github.com/garyburd/redigo/redis"
)

// Global variables
var (
	Config = struct {
		EmailServer      string
		EmailUsername    string
		EmailPassword    string
		EmailTestAddress string

		RedisDomain      string // = "tcp"
		RedisAddress     string // = "localhost:6379"
		RedisMaxIdle     int    // = 3
		RedisIdleTimeout int    // = 240

		WorkingDirectory    string
		SendMessageEndpoint string
	}{}
	Pool *redis.Pool
)

func init() {
	http.Handle("/assets/", http.FileServer(http.Dir("pages")))
	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/account/accept", accountAcceptHandler)
	http.HandleFunc("/account/activate", accountActivateHandler)
	http.HandleFunc("/account/delete", accountDeleteHandler)
	http.HandleFunc("/account/deny", accountDenyHandler)
	http.HandleFunc("/data/view", dataViewHandler)
	http.HandleFunc("/data/edit", dataEditHandler)
	http.HandleFunc("/isNameTaken", isNameTakenHandler)
	http.HandleFunc("/sendMessage", sendMessageHandler)
	http.HandleFunc("/signin", signinHandler)
	http.HandleFunc("/signout", signoutHandler)
	http.HandleFunc("/signup", signupHandler)

	// Create a thread-safe connection pool for redis
	Pool = &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial(Config.RedisDomain, Config.RedisAddress)
			if err != nil {
				return nil, err
			}

			return c, err
		},
	}
}

func readConfigFile(filepath string) error {
	f, err := os.Open(filepath)
	if err != nil {
		log.Fatal("main: os.Open:", err) // TODO handle file not found
		return err
	}

	r := csv.NewReader(f)
	r.Comma = ':'
	r.Comment = '#'
	r.FieldsPerRecord = 2
	r.LazyQuotes = true
	r.TrimLeadingSpace = true

	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal("main:", err)
		}

		switch record[0] {
		case "redis_domain":
			Config.RedisDomain = record[1]
			break
		case "redis_address":
			Config.RedisAddress = record[1]
			break
		case "redis_max_idle":
			i, err := strconv.Atoi(record[1])
			if err != nil {
				fmt.Printf("redis_max_idle value '%s' not valid. Using default.\n", record[1])
			} else {
				Config.RedisMaxIdle = i
			}
			break
		case "redis_idle_timeout":
			i, err := strconv.Atoi(record[1])
			if err != nil {
				fmt.Printf("redis_idle_timeout value '%s' not valid. Using default.\n", record[1])
			} else {
				Config.RedisIdleTimeout = i
			}
			break
		case "email_username":
			true, err := regexp.Match(`^.+@.+\..{2,}$`, []byte(record[1]))
			if !true {
				fmt.Println("The email", record[1], "is not valid.")
				return nil
			}
			if err != nil {
				return err
			}
			Config.EmailUsername = record[1]
			break
		case "email_password":
			Config.EmailPassword = record[1]
			break
		case "email_server":
			Config.EmailServer = record[1]
			break
		case "email_test_address":
			Config.EmailTestAddress = record[1]
			ok, err := regexp.Match(`^.+@.+\..{2,}$`, []byte(Config.EmailTestAddress))
			if !ok || err != nil {
				fmt.Println("The test email address provided isn't valid:", Config.EmailTestAddress)
				return nil
			}
			break
		case "working_directory":
			Config.WorkingDirectory = record[1]
			break
		case "bot_URI":
			Config.SendMessageEndpoint = record[1]
			break
		default:
			fmt.Printf("Parameter '%s' in config file not valid. Ignored.\n", record[0])
			break
		}
	}

	return err
}

func main() {
	// TODO get the config file from the command line
	if readConfigFile("/Users/marcofelici/go/src/webapp/webapp.conf") != nil {
		return
	}

	err := os.Chdir(Config.WorkingDirectory)
	if err != nil {
		log.Fatal("main: os.Chdir:", err)
		return
	}

	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("main: http.ListenAndServe:", err)
		return
	}
}
