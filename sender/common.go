package sender

import (
	b64 "encoding/base64"
	eventhub "github.com/Azure/azure-event-hubs-go/v3"
	"github.com/google/uuid"
	"math/rand"
	"strings"
)

const mLetterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func saltRandString(size int) string {
	randBytes := make([]byte, size)

	for i := range randBytes {
		randBytes[i] = mLetterBytes[rand.Intn(52)] //52 is the amount of characters in mLetterBytes
	}

	return string(randBytes)
}

func addProperties(event *eventhub.Event, properties []string) {
	var keyValue []string

	if event.Properties == nil {
		event.Properties = make(map[string]interface{})
	}

	for _, value := range properties {
		keyValue = strings.Split(value, ";")

		for _, v := range keyValue {
			keyValue = strings.Split(v, ":")
			if len(keyValue) == 2 {
				event.Properties[keyValue[0]] = keyValue[1]
			}
		}
	}
}

func createAnEvent(base64 bool, message string, withSuffix bool) *eventhub.Event {
	var event *eventhub.Event

	if withSuffix {
		message = message + "-" + saltRandString(12)
	}

	if base64 {
		decoded, err := b64.StdEncoding.DecodeString(message)
		if err == nil {
			event = eventhub.NewEvent(decoded)
		} else {
			event = eventhub.NewEvent([]byte(message))
		}
	} else {
		event = eventhub.NewEvent([]byte(message))
	}

	return event
}

func calcBatchLimit(sender *Sender, message string, withSuffix bool) (int, error) {
	event := createAnEvent(sender.base64String, message, withSuffix)
	addProperties(event, sender.properties)

	return getBatchLimit(event)
}

func calcBatchLimitWithEvents(events *[]*eventhub.Event) (int, error) {
	var biggestSize uint32 = 0
	var biggestEvent *eventhub.Event

	for _, event := range *events {
		size, err := getEventSize(event)
		if err == nil {
			if size > biggestSize {
				biggestEvent = event
				biggestSize = size
			}
		}
	}

	return getBatchLimit(biggestEvent)
}

func getEventSize(event *eventhub.Event) (uint32, error) {
	id := uuid.New()
	eb := eventhub.NewEventBatch(id.String(), nil)
	sizeBeforeEvent := eb.Size()

	_, err := eb.Add(event)

	if err == nil {
		msgSize := uint32(eb.Size() - sizeBeforeEvent)
		defer func() {
			eb.Clear()
			eb = nil
		}()
		return msgSize, nil
	}

	return 0, err
}

func getBatchLimit(event *eventhub.Event) (int, error) {
	id := uuid.New()
	eb := eventhub.NewEventBatch(id.String(), nil)
	sizeBeforeEvent := eb.Size()
	_, err := eb.Add(event)

	if err == nil {
		msgSize := eb.Size() - sizeBeforeEvent
		defer func() {
			eb.Clear()
			eb = nil
		}()
		return (int(eb.MaxSize) - 100) / msgSize, nil
	}

	return 0, err
}