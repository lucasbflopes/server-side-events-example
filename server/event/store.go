package event

import (
	"sync"
	"sync/atomic"
)

// Store stores subscribers and notify them when a new event is dispatched.
type Store struct {
	subscribersLock *sync.RWMutex
	id              uint64
	subscribers     map[uint64]Subscriber
}

// NewStore returns a new instance of a Store
func NewStore() *Store {
	return &Store{
		subscribersLock: &sync.RWMutex{},
		subscribers:     make(map[uint64]Subscriber),
	}
}

// Subscribe adds a new subscriber to the store.
// It returns a function that when called unsubscribes the subscriber from the store.
func (e *Store) Subscribe(subscriber Subscriber) func() {
	id := e.nextID()

	e.addSubscriber(id, subscriber)

	return func() { e.removeSubscriber(id) }
}

// Dispatch sends a new event to the store, broadcasting to all subscribers.
func (e *Store) Dispatch(event Event) {
	e.subscribersLock.RLock()
	defer e.subscribersLock.RUnlock()

	for _, o := range e.subscribers {
		o.Publish(event)
	}
}

// Subscriptions return the number of subscribers to the store.
func (e *Store) Subscriptions() int {
	e.subscribersLock.RLock()
	defer e.subscribersLock.RUnlock()

	return len(e.subscribers)
}

func (e *Store) nextID() uint64 {
	return atomic.AddUint64(&e.id, 1)
}

func (e *Store) addSubscriber(id uint64, subscriber Subscriber) {
	e.subscribersLock.Lock()
	defer e.subscribersLock.Unlock()

	e.subscribers[id] = subscriber
}

func (e *Store) removeSubscriber(id uint64) {
	e.subscribersLock.Lock()
	defer e.subscribersLock.Unlock()

	delete(e.subscribers, id)
}
