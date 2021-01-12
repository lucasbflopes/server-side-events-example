package event

const maxSubscriberBufferSize = 100

// Subscriber is an entity throught which a client can publish and consume events asynchronously.
type Subscriber interface {
	Publish(Event)
	EventStream() <-chan Event
}

// SubscriptionRule is a way to tell a subscriber whether it should consider processing an event or not.
type SubscriptionRule func(Event) bool

type subscriber struct {
	notificationChannel chan Event
	rules               []SubscriptionRule
}

// NewSubscriber returns a new Subscriber instance.
// To filter out undesired events, subscription rules can be supplied.
func NewSubscriber(rules ...SubscriptionRule) Subscriber {
	return &subscriber{
		notificationChannel: make(chan Event, maxSubscriberBufferSize),
		rules:               rules,
	}
}

// EventStream returns a channel throught which events published to the subscriber can be consumed asynchronously.
func (o *subscriber) EventStream() <-chan Event {
	return o.notificationChannel
}

// Publish publishs an event to the subscriber's event stream
// It ignores events which fails any subscription rule
func (o *subscriber) Publish(event Event) {
	for _, rule := range o.rules {
		if !rule(event) {
			return
		}
	}

	select {
	case o.notificationChannel <- event:
	default:
		// If the observer cannot process the message right away then we drop it to avoid backpressure.
	}

}
