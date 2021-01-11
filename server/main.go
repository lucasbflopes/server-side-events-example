package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

type eventStore struct {
	subLock *sync.Mutex
	id int
	observers map[int]chan<- string
}

func newEventStore() *eventStore {
	return &eventStore{
		subLock: &sync.Mutex{},
		observers: make(map[int]chan<- string),
	}
}

func (e *eventStore) Subscribe(channel chan<- string) func() {
	e.subLock.Lock()
	defer e.subLock.Unlock()

	e.observers[e.id] = channel

	result := func(m map[int]chan<- string, id int) func() {
		return func() {
			delete(m, id)
		}
	}(e.observers, e.id)

	e.id++

	return result
}

func (e *eventStore) Dispatch(message string) {
	for _, c := range e.observers {
		select {
			case c <- message:
			default:
		}
	}
}

type server struct {
	eventStore *eventStore
}

func (s server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	mux := http.NewServeMux()

	mux.HandleFunc("/sse/subscribe", s.handleSubscribe)
	mux.HandleFunc("/sse/dispatch", s.handleDispatch)

	mux.ServeHTTP(w, r)
}

func (s server) handleSubscribe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Add("Access-Control-Allow-Origin", "*")
	w.Header().Add("Content-Type", "text/event-stream")

	msgs := make(chan string)
	unsubscribe := s.eventStore.Subscribe(msgs)

processing:
	for {
		select {
			case <-r.Context().Done():
				unsubscribe()
				break processing
			case m := <-msgs:
				fmt.Fprint(w,  "event: dispatched\n")
				fmt.Fprintf(w, "data: %s\n\n", m)
			case <-time.After(30 * time.Second):
				fmt.Fprint(w, ": keepalive\n\n")
		}
		flusher.Flush()
	}
}

func (s server) handleDispatch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	data := r.URL.Query().Get("data")

	if data == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "`data` value is empty or missing from query string")
		return
	}

	s.eventStore.Dispatch(data)
}

func main() {
	s := server{ eventStore: newEventStore() }

	go func() {
		for {
			time.Sleep(500 * time.Millisecond)
			fmt.Printf("\rCurrent number of observers: %d", len(s.eventStore.observers))
		}
	}()

	log.Fatal(http.ListenAndServe(":8080", s))
}