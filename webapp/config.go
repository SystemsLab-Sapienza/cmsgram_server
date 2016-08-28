package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strconv"
)

var (
	config = struct {
		EmailServer      string
		EmailUsername    string
		EmailPassword    string
		EmailTestAddress string

		RedisDomain      string
		RedisAddress     string
		RedisMaxIdle     int
		RedisIdleTimeout int

		Domain              string
		MessageSendEndpoint string
		WorkingDirectory    string
	}{}
)

// Set the default configuration
func init() {
	config.RedisDomain = "tcp"
	config.RedisAddress = "localhost:6379"
	config.RedisMaxIdle = 3
	config.RedisIdleTimeout = 240

	config.Domain = "http://localhost:8080"
	config.WorkingDirectory = "/usr/local/bin"
}

func readConfigFile(filepath string) error {
	f, err := os.Open(filepath)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

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
			log.Fatal("readConfigFile():", err)
		}

		value := record[1]
		switch record[0] {
		case "domain":
			config.Domain = value
		case "redis_domain":
			config.RedisDomain = value
		case "redis_address":
			config.RedisAddress = value
		case "redis_max_idle":
			i, err := strconv.Atoi(value)
			if err != nil {
				fmt.Printf("redis_max_idle value '%s' not valid. Using default.\n", value)
			}

			config.RedisMaxIdle = i
		case "redis_idle_timeout":
			i, err := strconv.Atoi(value)
			if err != nil {
				fmt.Printf("redis_idle_timeout value '%s' not valid. Using default.\n", value)
			}

			config.RedisIdleTimeout = i
		case "email_username":
			true, err := regexp.Match(`^.+@.+\..{2,}$`, []byte(value))
			if !true {
				fmt.Println("The email", value, "is not valid.")
				return nil
			}
			if err != nil {
				return err
			}

			config.EmailUsername = value
		case "email_password":
			config.EmailPassword = value
		case "email_server":
			config.EmailServer = value
		case "email_test_address":
			ok, err := regexp.Match(`^.+@.+\..{2,}$`, []byte(value))
			if !ok || err != nil {
				fmt.Println("The test email address provided isn't valid:", value)
				return nil
			}

			config.EmailTestAddress = value
		case "working_directory":
			config.WorkingDirectory = value
		case "bot_URI":
			config.MessageSendEndpoint = value
		default:
			fmt.Printf("Parameter '%s' in config file not valid. Ignored.\n", record[0])
		}
	}

	return err
}
