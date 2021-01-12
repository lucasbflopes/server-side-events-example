package event

import (
	"sync"
	"sync/atomic"
)

// Store stores observers and notify them when a new event is dispatched
type Store struct {
	observerLock *sync.RWMutex
	id           uint64
	observers    map[uint64]chan<- string
}

// NewStore returns a new instance of a Store
func NewStore() *Store {
	return &Store{
		observerLock: &sync.RWMutex{},
		observers:    make(map[uint64]chan<- string),
	}
}

// Subscribe adds a new observer to the store through a channel to be notified.
// It returns a function that when called unsubscribes the observer from the store.
func (e *Store) Subscribe(channel chan<- string) func() {
	id := e.nextID()

	e.addObserver(id, channel)

	return func() { e.removeObserver(id) }
}

// Dispatch sends a new event to the store, broadcasting to all observers
func (e *Store) Dispatch(message string) {
	e.observerLock.RLock()
	defer e.observerLock.RUnlock()

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
	return e.countObservers()
}

func (e *Store) nextID() uint64 {
	return atomic.AddUint64(&e.id, 1)
}

func (e *Store) addObserver(id uint64, channel chan<- string) {
	e.observerLock.Lock()
	defer e.observerLock.Unlock()

	e.observers[id] = channel
}

func (e *Store) removeObserver(id uint64) {
	e.observerLock.Lock()
	defer e.observerLock.Unlock()

	delete(e.observers, id)
}

func (e *Store) countObservers() int {
	e.observerLock.RLock()
	defer e.observerLock.RUnlock()

	return len(e.observers)
}
