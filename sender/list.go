package sender

import (
	eventhub "github.com/Azure/azure-event-hubs-go/v3"
)

// List interface that all lists implement
type EventList interface {
	Get(index int) (*eventhub.Event, bool)
	Remove(index int)
	Add(values ...*eventhub.Event)
	Contains(values ...*eventhub.Event) bool
	Swap(index1, index2 int)
	Insert(index int, values ...*eventhub.Event)
	Set(index int, value *eventhub.Event)

	Container
	// Empty() bool
	// Size() int
	// Clear()
	// Values() []interface{}
}

type Container interface {
	Empty() bool
	Size() int
	Clear()
	Values() []*eventhub.Event
}