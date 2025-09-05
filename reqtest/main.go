package main

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

var respDelay = flag.Duration("resp-delay", 0*time.Second, "Adds a delay before responding to a request in listen mode")

func main() {
	flag.Parse()
	args := flag.Args()
	if len(args) == 0 {
		printUsage()
		os.Exit(1)
	}

	switch args[0] {
	case "listen":
		if err := listen(args[1:]); err != nil {
			log.Printf("failed to listen: %v\n", err)
			os.Exit(1)
		}
	case "send":
		if err := send(args[1:]); err != nil {
			log.Printf("failed to send: %v\n", err)
			os.Exit(1)
		}
	default:
		log.Printf("unknown arg %v", args[0])
		printUsage()
		os.Exit(1)
	}

	os.Exit(0)
}

func printUsage() {
	fmt.Println(`This is a tool that will listen for any requests and echo them to stdout.
It can also send requests of increasing sizes to a listener.

To listen:
[binary] listen <address>

To send:
[binary] send <address>`)
}

func listen(args []string) error {
	if len(args) != 1 {
		printUsage()
		return errors.New("listen expects exactly 1 argument")
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Println("received request")
		if respDelay != nil && *respDelay > 0*time.Second {
			log.Printf("waiting %s before reading/responding...", *respDelay)
			time.Sleep(*respDelay)
		}

		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			log.Printf("error reading body: %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		log.Printf("read %v bytes from body\n", len(bodyBytes))
	})

	log.Printf("listening on %v\n", args[0])
	return http.ListenAndServe(args[0], nil)
}

func send(args []string) error {
	if len(args) != 1 {
		printUsage()
		return errors.New("send expects exactly 1 argument")
	}

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	maxBytes := 1 << 25 // max out at 32M
	bytesToSend := 1 << 2
	for bytesToSend <= maxBytes {
		log.Printf("sending %v bytes\n", bytesToSend)
		b := make([]byte, bytesToSend/2)
		if _, err := rand.Read(b); err != nil {
			return fmt.Errorf("failed to generate bytes: %w", err)
		}

		bodyStr := hex.EncodeToString(b)
		req, err := http.NewRequest("PUT", args[0], bytes.NewReader([]byte(bodyStr)))
		if err != nil {
			return fmt.Errorf("could make request: %w", err)
		}

		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("could not execute request: %w", err)
		}

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("did not get 200 response, got %v", resp.StatusCode)
		}

		bytesToSend <<= 1
	}

	return nil
}
