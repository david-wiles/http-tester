package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"time"
)

type Message struct {
	timestamp time.Time
	text      string
	error     bool
}

func main() {

	// Test options
	delayPtr := flag.Int64("delay", 500, "Maximum delay between successive requests in ms")
	maxErrorPtr := flag.Int("max-errors", 100, "Max number of successive errors before terminating program")
	lengthLimitPtr := flag.Int("log-length", 1000, "Maximum length of response body to print")

	// Request options
	urlPtr := flag.String("url", "localhost", "The url to make requests to")
	methodPtr := flag.String("method", "GET", "The http method to use in the request")
	contentTypePtr := flag.String("content-type", "", "Content-Type header")
	authPtr := flag.String("auth", "", "The auth header to use in the request.")
	srcPtr := flag.String("src", "", "The data source for requests. Valid options are novel. This is used for the query "+
		"string in GET requests and the request body in POST requests")
	filePtr := flag.String("file", "", "Input file for file source types")

	flag.Parse()

	// Copy configuration values used in program
	delay := *delayPtr
	maxError := *maxErrorPtr
	lengthLimit := *lengthLimitPtr
	url := *urlPtr
	method := *methodPtr
	auth := *authPtr
	contentType := *contentTypePtr

	stream, err := GetInputStream(*srcPtr, *filePtr)
	if err != nil {
		fmt.Println("Could not create input stream from", *srcPtr, *filePtr, err.Error())
		return
	}

	messages := make(chan Message)

	// Consume messages
	go func() {
		// Count successive errors to terminate program after unrecoverable errors
		errorCount := 0

		for {
			message := <-messages

			if message.error {
				errorCount += 1
				fmt.Println(message.timestamp.Format(time.RFC3339Nano), "ERROR:", message.text)
			} else {
				errorCount = 0
				fmt.Println(message.timestamp.Format(time.RFC3339Nano), message.text)
			}

			if errorCount > maxError {
				panic("Too many errors. Exiting")
			}
		}
	}()

	for {

		if delay > 0 {
			// Sleep for a random amount of time between 0 and the maximum delay
			time.Sleep(time.Millisecond * time.Duration(rand.Int63n(delay)))
		}

		// Produce messages to be printed to stdout
		go func() {

			var reqBody io.Reader = nil

			if method == "GET" || method == "get" {
				next := stream.Next()
				if next != nil {
					b, err := ioutil.ReadAll(next)
					if err != nil {
						messages <- Message{time.Now(), err.Error(), true}
						return
					}
					url += string(b)
				}
			} else {
				reqBody = stream.Next()
			}

			req, err := http.NewRequest(method, url, reqBody)
			if err != nil {
				messages <- Message{time.Now(), err.Error(), true}
				return
			}

			if auth != "" {
				req.Header.Set("Authorization", auth)
			}
			req.SetBasicAuth("david", "ABC!@#abc123")

			if contentType != "" {
				req.Header.Set("Content-Type", contentType)
			}

			start := time.Now()
			resp, err := http.DefaultClient.Do(req)
			duration := time.Since(start)

			if err != nil {
				messages <- Message{time.Now(), err.Error(), true}
				return
			}
			defer resp.Body.Close()

			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				messages <- Message{time.Now(), err.Error(), true}
				return
			}

			if len(body) > lengthLimit {
				body = body[:lengthLimit]
			}

			messages <- Message{
				time.Now(),
				fmt.Sprintf("Recieved response after %d ms. %s %s %d: %q", duration.Milliseconds(), method, url, resp.StatusCode, body),
				false,
			}
		}()
	}
}
