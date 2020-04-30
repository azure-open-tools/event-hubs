package receiver

import (
	"context"
	eventhub "github.com/Azure/azure-event-hubs-go/v3"
	"log"
	"strings"
	"time"
)

func listenToSpecificPartitions(ctx context.Context, receiver *Receiver) error {
	// listener to a single or multiple partition ids
	for _, partitionID := range receiver.partitionIds {
		_, err := receiver.eHub.Receive(ctx, partitionID, receiver.onReceive,
			eventhub.ReceiveFromTimestamp(time.Now()),
			eventhub.ReceiveWithConsumerGroup(receiver.consumerGroup))

		if err != nil {
			log.Println(err.Error())
			return err
		}
	}

	return nil
}

func listenToAllAvailablePartitions(ctx context.Context, receiver *Receiver) error {
	runtimeInfo, err := receiver.eHub.GetRuntimeInformation(ctx)

	if err != nil {
		log.Println(err.Error())
		return err
	}

	// listen to each partition of the Event Hub
	for _, partitionID := range runtimeInfo.PartitionIDs {
		_, err := receiver.eHub.Receive(ctx, partitionID, receiver.onReceive,
			eventhub.ReceiveFromTimestamp(time.Now()),
			eventhub.ReceiveWithConsumerGroup(receiver.consumerGroup))

		if err != nil {
			log.Println(err.Error())
			return err
		}
	}

	return nil
}

func checkDataFilter(event *eventhub.Event, receiver *Receiver) *eventhub.Event {
	data := string(event.Data)
	for i := 0; i < len(receiver.dataFilter); i++ {
		if strings.Contains(data, receiver.dataFilter[i]) {
			return event
		}
	}

	return nil
}

func checkPropertyFilter(event *eventhub.Event, receiver *Receiver) *eventhub.Event {
	var keyValue []string
	var sep = ":"

	for i := 0; i < len(receiver.propertyFilter); i++ {
		// check for a specific property: "propertyToSearch:valueToSearch" "<property-name>:<property-value>"
		keyValue = strings.Split(receiver.propertyFilter[i], sep)
		if len(keyValue) == 2 {
			value, ok := event.Properties[keyValue[0]]
			if ok &&  strings.Contains(value.(string), keyValue[1]) {
				return event
			}
		} else {
			// otherwise, search for all properties with some value: "value-to-search"
			for _, value := range event.Properties {
				if strings.Contains(value.(string), receiver.propertyFilter[i]) {
					return event
				}
			}
		}
	}

	return nil
}

