package event

import (
	"sync"
)

// Store stores observers and notify them when a new event is dispatched
type Store struct {
	subLock   *sync.Mutex
	id        int
	observers map[int]chan<- string
}

// NewStore returns a new instance of a Store
func NewStore() *Store {
	return &Store{
		subLock:   &sync.Mutex{},
		observers: make(map[int]chan<- string),
	}
}

// Subscribe adds a new observer to the store through a channel to be notified.
// It returns a function that when called unsubscribes the observer from the store.
func (e *Store) Subscribe(channel chan<- string) func() {
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

// Dispatch sends a new event to the store, broadcasting to all observers
func (e *Store) Dispatch(message string) {
	for _, c := range e.observers {
		select {
		case c <- message:
		default:
			// If the observer cannot process the message right away then we drop it to avoid backpressure.
		}
	}
}

// Subscriptions return the number of observers subscribed to the store
func (e *Store) Subscriptions() int {
	return len(e.observers)
}
