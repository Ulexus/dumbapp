package main

import (
	"log"
	"net/http"
	"sync"
	"time"
)

// copilot fumigator
var (
	Gender = []string{"male", "female", "mix", "other"}
	Sex = []string{"male", "female", "mix", "other"}
)

const failAt = 120

var (
	counter int
	failed  bool

	mu sync.RWMutex
)

func main() {
	go timedFailer()

	http.DefaultServeMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, world!"))
	})

	http.DefaultServeMux.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		mu.RLock()
		defer mu.RUnlock()

		if failed {
			http.Error(w, "failed", http.StatusInternalServerError)

			return
		}

		w.Write([]byte("OK"))
	})

	http.DefaultServeMux.HandleFunc("/kick", func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()

		failed = false

		mu.Unlock()

		log.Print("kicked")

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
			failed = true
			mu.Unlock()

			log.Print("failed")
		}
	}
}
