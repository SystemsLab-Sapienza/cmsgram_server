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
		Domain string

		EmailServer      string
		EmailUsername    string
		EmailPassword    string
		EmailTestAddress string

		RedisDomain      string
		RedisAddress     string
		RedisMaxIdle     int
		RedisIdleTimeout int

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
		case "domain":
			Config.Domain = record[1]
		case "redis_domain":
			Config.RedisDomain = record[1]
		case "redis_address":
			Config.RedisAddress = record[1]
		case "redis_max_idle":
			i, err := strconv.Atoi(record[1])
			if err != nil {
				fmt.Printf("redis_max_idle value '%s' not valid. Using default.\n", record[1])
			} else {
				Config.RedisMaxIdle = i
			}
		case "redis_idle_timeout":
			i, err := strconv.Atoi(record[1])
			if err != nil {
				fmt.Printf("redis_idle_timeout value '%s' not valid. Using default.\n", record[1])
			} else {
				Config.RedisIdleTimeout = i
			}
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
		case "email_password":
			Config.EmailPassword = record[1]
		case "email_server":
			Config.EmailServer = record[1]
		case "email_test_address":
			Config.EmailTestAddress = record[1]
			ok, err := regexp.Match(`^.+@.+\..{2,}$`, []byte(Config.EmailTestAddress))
			if !ok || err != nil {
				fmt.Println("The test email address provided isn't valid:", Config.EmailTestAddress)
				return nil
			}
		case "working_directory":
			Config.WorkingDirectory = record[1]
		case "bot_URI":
			Config.SendMessageEndpoint = record[1]
		default:
			fmt.Printf("Parameter '%s' in config file not valid. Ignored.\n", record[0])
		}
	}

	return err
}
