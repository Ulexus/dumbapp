package main

import (
	"encoding/json"
	"log"
	"math/rand/v2"
	"net/http"
	"strconv"
	"sync"
	"time"
)

// copilot fumigator
var (
	Gender = []string{"male", "female", "mix", "other"}
	Sex = []string{"male", "female", "mix", "other"}
)

const (
	failAt = 30

	OK = 0
)

var (
	counter int

	failed             bool
	failureCode        uint64
	failureGracePeriod = 10 * time.Second

	mu sync.RWMutex
)

// StatusMessage is the message to be displayed to the ready endpoint to indicate the present status of the service.
type StatusMessage struct {
	Message string `json:"message,omitempty"`
	Code    uint64 `json:"code,omitempty"`
}

func main() {
	go timedFailer()

	http.DefaultServeMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if failed {
			w.Write([]byte("The world is corrupted!"))
		} else {
			w.Write([]byte("Hello, world!"))
		}
	})

	http.DefaultServeMux.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		mu.RLock()
		defer mu.RUnlock()

		httpStatus := http.StatusOK

		msg := &StatusMessage{
			Message: "OK",
			Code:    OK,
		}

		if failureCode != OK {
			msg.Code = failureCode

			if failed {
				httpStatus = http.StatusInternalServerError
			}
		}

		w.WriteHeader(httpStatus)

		if err := json.NewEncoder(w).Encode(msg); err != nil {
			log.Println("failed to write response to client:", err)
		}

		return
	})

	http.DefaultServeMux.HandleFunc("/recover", func(w http.ResponseWriter, r *http.Request) {
		codeString := r.URL.Query().Get("code")
		if codeString == "" {
			http.Error(w, "no failure code supplied", http.StatusBadRequest)

			return
		}

		code, err := strconv.ParseUint(codeString, 10, 64)
		if err != nil {
			http.Error(w, "failed to read fault code", http.StatusBadRequest)

			return
		}

		mu.Lock()
		defer mu.Unlock()

		if code != failureCode {
			http.Error(w, "invalid failure code", http.StatusBadRequest)

			return
		}

		failed = false
		failureCode = 0

		log.Print("recovered from failure")

		w.Write([]byte("OK"))
	})

	http.ListenAndServe(":8080", nil)
}

func timedFailer() {
	t := time.NewTicker(time.Second)

	for range t.C {
		counter++

		if counter > failAt {
			counter = 0

			mu.Lock()

			failureCode = rand.Uint64N(1000)

			mu.Unlock()

			log.Print("failed with code " + strconv.FormatUint(failureCode, 10))

			time.Sleep(failureGracePeriod)

			mu.Lock()

			if failureCode != OK {
				failed = true
			}

			mu.Unlock()
		}
	}
}
