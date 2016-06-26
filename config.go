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
)

func readConfigFile(filepath string) error {
	f, err := os.Open(filepath)
	if err != nil {
		log.Fatal("main: os.Open:", err) // TODO handle file not found
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
