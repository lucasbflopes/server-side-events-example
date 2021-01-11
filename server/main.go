package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

var latestEventID = 0
var idMutex = &sync.Mutex{}

func main() {
	http.HandleFunc("/events", handleEvents)

	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleEvents(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Access-Control-Allow-Origin", "*")
	w.Header().Add("Content-Type", "text/event-stream")

processing:
	for {
		select {
			case <-r.Context().Done():
				break processing
			default:
				fmt.Printf("Event is going to be processed: %d\n", latestEventID)

				time.Sleep(1 * time.Second)
				fmt.Fprint(w, "event: foo\n")

				idMutex.Lock()
				fmt.Fprintf(w, "data: event #%d \n\n", latestEventID)
				latestEventID++
				idMutex.Unlock()

				f, ok := w.(http.Flusher)

				if !ok {
					log.Fatalf("Cannot flush response body")
				}

				f.Flush()
		}
	}
}
