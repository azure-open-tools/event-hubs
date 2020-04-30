package sender

import (
	"context"
	"errors"
	eventhub "github.com/Azure/azure-event-hubs-go/v3"
	"log"
	"runtime"
	"strings"
	"sync"
)

type (
	ISenderBuilder interface {
		AddPartitionId(partitionId string) ISenderBuilder
		AddPartitionIds(partitionIds []string) ISenderBuilder
		AddProperty(filter string) ISenderBuilder
		AddProperties(filters []string) ISenderBuilder
		SetBase64(is64base bool) ISenderBuilder
		SetNumberOfMessages(amount int64) ISenderBuilder
		SetRandomMessageSuffix(withSuffix bool) ISenderBuilder
		SetConnectionString(connStr string) ISenderBuilder
		SetOnAfterSendMessage(handler func(event *eventhub.Event)) ISenderBuilder
		SetOnBeforeSendMessage(handler func(event *eventhub.Event)) ISenderBuilder
		SetOnAfterSendBatchMessage(handler func(batchSizeSent int, workerIndex int)) ISenderBuilder
		SetOnBeforeSendBatchMessage(handler func(batchSize int, workerIndex int)) ISenderBuilder
		GetSender() (*Sender, error)
	}

	Builder struct {
		base64String             bool
		connString               string
		numberOfMessages         int64
		messageSuffix            bool
		partitionIds             []string
		properties               []string
		onAfterSendMessage       func(event *eventhub.Event)
		onBeforeSendMessage      func(event *eventhub.Event)
		onAfterSendBatchMessage  func(batchSizeSent int, workerIndex int)
		onBeforeSendBatchMessage func(batchSize int, workerIndex int)
	}

	ISender interface {
		AddProperties(properties map[string]interface{})
		SendMessage(message string, ctx context.Context) error
		SendBatchMessage(message string, ctx context.Context) error
	}

	Sender struct {
		// internal fields
		eHub             *eventhub.Hub
		base64String     bool
		connString       string
		numberOfMessages int64
		messageSuffix    bool
		partitionIds     []string
		properties       []string
		//onBatchesCreated         func(ctx context.Context, event *eventhub.Event) error
		onAfterSendMessage       func(event *eventhub.Event)
		onBeforeSendMessage      func(event *eventhub.Event)
		onAfterSendBatchMessage  func(batchSizeSent int, workerIndex int)
		onBeforeSendBatchMessage func(batchSize int, workerIndex int)
	}
)

func NewSenderBuilder() *Builder {
	return &Builder{}
}

func (builder *Builder) AddProperty(filter string) ISenderBuilder {
	if len(strings.TrimSpace(filter)) > 0 {
		builder.properties = append(builder.properties, filter)
	}

	return builder
}

func (builder *Builder) AddProperties(filters []string) ISenderBuilder {
	if len(filters) > 0 {
		builder.properties = append(builder.properties, filters...)
	}

	return builder
}

func (builder *Builder) AddPartitionId(partitionId string) ISenderBuilder {
	if len(strings.TrimSpace(partitionId)) > 0 {
		builder.partitionIds = append(builder.partitionIds, partitionId)
	}

	return builder
}

func (builder *Builder) AddPartitionIds(partitionIds []string) ISenderBuilder {
	if len(partitionIds) > 0 {
		builder.partitionIds = append(builder.partitionIds, partitionIds...)
	}

	return builder
}

func (builder *Builder) SetBase64(is64base bool) ISenderBuilder {
	builder.base64String = is64base

	return builder
}

func (builder *Builder) SetNumberOfMessages(amount int64) ISenderBuilder {
	if amount <= 0 {
		amount = 1
	}

	builder.numberOfMessages = amount

	return builder
}

func (builder *Builder) SetRandomMessageSuffix(withSuffix bool) ISenderBuilder{
	builder.messageSuffix = withSuffix

	return builder
}

func (builder *Builder) SetConnectionString(connStr string) ISenderBuilder {
	if len(strings.TrimSpace(connStr)) > 0 {
		builder.connString = connStr
	}

	return builder
}

func (builder *Builder) SetOnAfterSendMessage(handler func(event *eventhub.Event)) ISenderBuilder {
	if handler != nil {
		builder.onAfterSendMessage = handler
	}

	return builder
}

func (builder *Builder) SetOnBeforeSendMessage(handler func(event *eventhub.Event)) ISenderBuilder {
	if handler != nil {
		builder.onBeforeSendMessage = handler
	}

	return builder
}

func (builder *Builder) SetOnAfterSendBatchMessage(handler func(batchSizeSent int, workerIndex int)) ISenderBuilder {
	if handler != nil {
		builder.onAfterSendBatchMessage = handler
	}

	return builder
}

func (builder *Builder) SetOnBeforeSendBatchMessage(handler func(batchSize int, workerIndex int)) ISenderBuilder {
	if handler != nil {
		builder.onBeforeSendBatchMessage = handler
	}

	return builder
}

func (builder *Builder) GetSender() (*Sender, error) {
	if len(strings.TrimSpace(builder.connString)) == 0 {
		return nil, errors.New("connection string is missing")
	}

	sender := &Sender{}
	sender.base64String = builder.base64String
	sender.connString = builder.connString
	sender.numberOfMessages = builder.numberOfMessages
	sender.messageSuffix = builder.messageSuffix
	sender.partitionIds = builder.partitionIds
	sender.properties =  builder.properties
	sender.onAfterSendMessage = builder.onAfterSendMessage
	sender.onBeforeSendMessage = builder.onBeforeSendMessage
	sender.onAfterSendBatchMessage = builder.onAfterSendBatchMessage
	sender.onBeforeSendBatchMessage = builder.onBeforeSendBatchMessage

	hub, err := eventhub.NewHubFromConnectionString(sender.connString)
	if err == nil {
		sender.eHub = hub
	}

	return sender, nil
}

func (sender* Sender) SendMessage(message string, ctx context.Context) error {
	var err error = nil
	var i int64

	for i = 0; i < sender.numberOfMessages; i++ {
		event := createAnEvent(sender.base64String, message, sender.messageSuffix)
		addProperties(event, sender.properties)

		if sender.onBeforeSendMessage != nil {
			sender.onBeforeSendMessage(event)
		}

		if err = sendMessage(sender.eHub, ctx, event); err != nil {
			break
		}

		if sender.onAfterSendMessage != nil {
			sender.onAfterSendMessage(event)
		}
	}

	return err
}

func (sender* Sender) SendBatchMessage(message string, ctx context.Context) error {
	limit, err := calcBatchLimit(sender, message, sender.messageSuffix)

	if err == nil {
		numGoRoutines := runtime.NumCPU()
		eventBatches := createEventBatchCollection(sender, numGoRoutines, int64(limit), sender.numberOfMessages,
			message, sender.messageSuffix)
		if len(eventBatches) > 0 {
			var wg sync.WaitGroup

			sender.triggerBatches(ctx, &wg, numGoRoutines, eventBatches)
		}
	}

	return err
}

func (sender *Sender) AddProperties(properties map[string]interface{}) {
	if len(properties) > 0 {
		var entry string

		for key, value := range properties {
			entry = key + ":" + value.(string)
			sender.properties = append(sender.properties, entry)
		}
	}
}

func sendMessage(hub* eventhub.Hub, ctx context.Context, event* eventhub.Event) error {
	defer func () {
		err := hub.Close(ctx)
		if err != nil {
			log.Println("error on close event hub connection: ", err)
		}
	} ()

	return hub.Send(ctx, event)
}