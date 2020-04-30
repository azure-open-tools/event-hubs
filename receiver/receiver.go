package receiver

import (
	"context"
	"errors"
	eventhub "github.com/Azure/azure-event-hubs-go/v3"
	"strings"
)

type (
	IReceiveBuilder interface {
		AddDataFilter(filter string) IReceiveBuilder
		AddDataFilters(filters []string) IReceiveBuilder
		AddPropertyFilter(filter string) IReceiveBuilder
		AddPropertyFilters(filters []string) IReceiveBuilder
		AddListenerPartitionId(partitionId string) IReceiveBuilder
		AddListenerPartitionIds(partitionIds []string) IReceiveBuilder
		SetConnectionString(connStr string) IReceiveBuilder
		SetConsumerGroup(consumerGroup string) IReceiveBuilder
		SetReceiverHandler(handler func(ctx context.Context, event *eventhub.Event) error) IReceiveBuilder
		GetReceiver() (*Receiver, error)
	}

	IReceiver interface {
		StartListener(ctx context.Context) error
		StopListener(ctx context.Context) error
	}

	Receiver struct {
		// internal fields
		dataFilter       []string
		propertyFilter   []string
		partitionIds     []string
		connString       string
		consumerGroup    string
		onReceiveHandler func(ctx context.Context, event *eventhub.Event) error

		eHub           *eventhub.Hub
	}

	Builder struct {
		DataFilter       []string
		PropertyFilter   []string
		PartitionIds     []string
		ConnString       string
		ConsumerGroup    string

		OnReceiveHandler func(ctx context.Context, event *eventhub.Event) error
	}
)

func NewReceiverBuilder() *Builder {
	return &Builder{}
}

func (builder *Builder) AddDataFilter(filter string) IReceiveBuilder {
	if len(strings.TrimSpace(filter)) > 0 {
		builder.DataFilter = append(builder.DataFilter, filter)
	}

	return builder
}

func (builder *Builder) AddDataFilters(filters []string) IReceiveBuilder {
	if len(filters) > 0 {
		builder.DataFilter = append(builder.DataFilter, filters...)
	}

	return builder
}

func (builder *Builder) AddPropertyFilter(filter string) IReceiveBuilder {
	if len(strings.TrimSpace(filter)) > 0 {
		builder.PropertyFilter = append(builder.PropertyFilter, filter)
	}

	return builder
}

func (builder *Builder) AddPropertyFilters(filters []string) IReceiveBuilder {
	if len(filters) > 0 {
		builder.PropertyFilter = append(builder.PropertyFilter, filters...)
	}

	return builder
}

func (builder *Builder) AddListenerPartitionId(partitionId string) IReceiveBuilder {
	if len(strings.TrimSpace(partitionId)) > 0 {
		builder.PartitionIds = append(builder.PartitionIds, partitionId)
	}

	return builder
}

func (builder *Builder) AddListenerPartitionIds(partitionIds []string) IReceiveBuilder {
	if len(partitionIds) > 0 {
		builder.PartitionIds = append(builder.PartitionIds, partitionIds...)
	}

	return builder
}

func (builder *Builder) SetConnectionString(connStr string) IReceiveBuilder {
	if len(strings.TrimSpace(connStr)) > 0 {
		builder.ConnString = connStr
	}

	return builder
}

func (builder *Builder) SetConsumerGroup(consumerGroup string) IReceiveBuilder {
	if len(strings.TrimSpace(consumerGroup)) > 0 {
		builder.ConsumerGroup = consumerGroup
	}

	return builder
}

func (builder *Builder) SetReceiverHandler(handler func(ctx context.Context, event *eventhub.Event) error) IReceiveBuilder {
	if handler != nil {
		builder.OnReceiveHandler = handler
	}

	return builder
}

func (builder *Builder) GetReceiver() (*Receiver, error) {
	if len(strings.TrimSpace(builder.ConnString)) == 0 {
		return nil, errors.New("connection string is missing")
	}

	receiver := &Receiver{}
	receiver.connString = builder.ConnString

	if len(strings.TrimSpace(builder.ConsumerGroup)) > 0 {
		receiver.consumerGroup = builder.ConsumerGroup
	} else {
		receiver.consumerGroup = eventhub.DefaultConsumerGroup
	}

	receiver.dataFilter = builder.DataFilter
	receiver.propertyFilter = builder.PropertyFilter
	receiver.partitionIds = builder.PartitionIds
	receiver.onReceiveHandler = builder.OnReceiveHandler

	hub, err := eventhub.NewHubFromConnectionString(receiver.connString)
	if err == nil {
		receiver.eHub = hub
	}

	return receiver, err
}

func (receiver *Receiver) StartListener(ctx context.Context) error {
	var err error

	if len(receiver.partitionIds) > 0 {
		err = listenToSpecificPartitions(ctx, receiver)
	} else {
		err = listenToAllAvailablePartitions(ctx, receiver)
	}

	return err
}

func (receiver *Receiver) StopListener(ctx context.Context) error {
	if receiver.eHub != nil {
		return receiver.eHub.Close(ctx)
	}

	return nil
}

func (receiver *Receiver) onReceive(ctx context.Context, event *eventhub.Event) error {
	if len(receiver.dataFilter) > 0 {
		evt := checkDataFilter(event, receiver)
		if evt != nil {
			return receiver.onReceiveHandler(ctx, evt)
		}
	}

	if len(receiver.propertyFilter) > 0 {
		evt := checkPropertyFilter(event, receiver)
		if evt != nil {
			return receiver.onReceiveHandler(ctx, evt)
		}
	} else {
		return receiver.onReceiveHandler(ctx, event)
	}

	return nil
}