package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
)

var (
	serverURL string
	authToken string
)

func main() {
	// Global flags (must be before subcommand)
	// Usage: mailraven-cli -server=... -token=... users list
	flag.StringVar(&serverURL, "server", "http://localhost:8443", "MailRaven API URL")
	flag.StringVar(&authToken, "token", "", "Admin Auth Token (or MAILRAVEN_ADMIN_TOKEN env)")

	flag.Parse()

	if authToken == "" {
		authToken = os.Getenv("MAILRAVEN_ADMIN_TOKEN")
	}

	if len(flag.Args()) < 1 {
		usage()
		os.Exit(1)
	}

	command := flag.Arg(0)
	subArgs := flag.Args()[1:]

	switch command {
	case "users":
		handleUsers(subArgs)
	case "system":
		handleSystem(subArgs)
	default:
		fmt.Printf("Unknown command: %s\n", command)
		usage()
		os.Exit(1)
	}
}

func usage() {
	fmt.Println("Usage: mailraven-cli [flags] <command> <subcommand> [args]")
	fmt.Println("Commands: users, system")
	fmt.Println("Flags:")
	flag.PrintDefaults()
}

// Client helper
func apiRequest(method, path string, body interface{}) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequest(method, serverURL+"/api/v1/admin"+path, bodyReader)
	if err != nil {
		return nil, err
	}

	if authToken != "" {
		req.Header.Set("Authorization", "Bearer "+authToken)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	return client.Do(req)
}
