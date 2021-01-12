package event

// Event represents an event that can be publish and consumed
type Event struct {
	Name    string
	Payload string
}
