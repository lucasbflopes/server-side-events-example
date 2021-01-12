package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/lucasbflopes/server-side-events-example/event"
)

type server struct {
	eventStore *event.Store
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

	subscriber := event.NewSubscriber()
	unsubscribe := s.eventStore.Subscribe(subscriber)

processing:
	for {
		select {
		case <-r.Context().Done():
			unsubscribe()
			break processing
		case m := <-subscriber.EventStream():
			fmt.Fprint(w, "event: dispatched\n")
			fmt.Fprintf(w, "data: %s\n\n", m.Payload)
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

	s.eventStore.Dispatch(event.Event{Payload: data})
}

func main() {
	s := server{
		eventStore: event.NewStore(),
	}

	go func() {
		for {
			time.Sleep(500 * time.Millisecond)
			fmt.Printf("\rCurrent number of observers: %d", s.eventStore.Subscriptions())
		}
	}()

	log.Fatal(http.ListenAndServe(":8080", s))
}
