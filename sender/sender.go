// This package helps you to send messages(events) to Azure Event Hubs.
//
// you can send message, send messages in batch  which it allows you to send more events at once (it depends of the listSize of your message/event content and properties).
// you also can enrich your event with metadata adding properties and handle the event before or after send with event handlers.
//
// See more about Azure Event Hubs
//
// https://azure.microsoft.com/services/event-hubs
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
	// ISenderBuilder this interface defines all properties and handlers available to this package functionalities.
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

	// Builder struct implements ISenderBuilder which handles sender creation.
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

	// ISender defines methods to send operations against azure event hubs.
	ISender interface {
		AddProperties(properties map[string]interface{})
		SendMessage(message string, ctx context.Context) error
		SendBatchMessage(message string, ctx context.Context) error
		SendEventsAsBatch(events *[]*eventhub.Event) error
	}

	// Sender struct implements ISender interface methods.
	Sender struct {
		//internal fields
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

// NewSenderBuilder() creates instance of Builder.
func NewSenderBuilder() *Builder {
	return &Builder{}
}

// AddProperty(filter string) add a single property to the event, the format expected is: "propertyKey:propertyValue"
func (builder *Builder) AddProperty(filter string) ISenderBuilder {
	if len(strings.TrimSpace(filter)) > 0 {
		builder.properties = append(builder.properties, filter)
	}

	return builder
}

// AddProperty(filters []string) add a slice of properties ex: []string{"propertyKey1:propertyValue1", "propertyKey2:propertyValue2"}
func (builder *Builder) AddProperties(filters []string) ISenderBuilder {
	if len(filters) > 0 {
		builder.properties = append(builder.properties, filters...)
	}

	return builder
}

// AddPartitionId(partitionId string) add single partition. Format expected: "0" (a integer among 0 to 32, it will depends of your Event Hubs settings)
func (builder *Builder) AddPartitionId(partitionId string) ISenderBuilder {
	if len(strings.TrimSpace(partitionId)) > 0 {
		builder.partitionIds = append(builder.partitionIds, partitionId)
	}

	return builder
}

// AddPartitionIds(partitionIds []string) add a slice of partitionIds ex: []string{"0", "1", ..., "32"}
func (builder *Builder) AddPartitionIds(partitionIds []string) ISenderBuilder {
	if len(partitionIds) > 0 {
		builder.partitionIds = append(builder.partitionIds, partitionIds...)
	}

	return builder
}

// SetBase64(isBase64 bool) set true in case of your message is a binary content.
// it needs to the sender convert back to the binary format([]byte) before send.
func (builder *Builder) SetBase64(is64base bool) ISenderBuilder {
	builder.base64String = is64base

	return builder
}

// SetNumberOfMessages(amount int64) set the number of messages you want to send. This will repeat the message content
// and send them according the amount you have set.
func (builder *Builder) SetNumberOfMessages(amount int64) ISenderBuilder {
	if amount <= 0 {
		amount = 1
	}

	builder.numberOfMessages = amount

	return builder
}

// SetRandomMessageSuffix(withSuffix bool) to avoid the message repetition content in usage along side with SetNumberOfMessages
// with suffix the library will generate 12 random characters and append in the end of the message content. It useful
// to make sure different messages are being sent to the event hubs. the format delivered will be something like:
// "message123-asuYRadaIY1d, where "-asuYRadaIY1d" is the suffix random content generated.
func (builder *Builder) SetRandomMessageSuffix(withSuffix bool) ISenderBuilder{
	builder.messageSuffix = withSuffix

	return builder
}

// SetConnectionString(connStr string) Required field. Connection string format is something like:
// "Endpoint=sb://<namespace>.servicebus.windows.net/;SharedAccessKeyName=send;SharedAccessKey=<AccessKey>;EntityPath=<topic>'""
// note SharedAccessKeyName, there are two available values possible: send and listen. To this library must be SEND
func (builder *Builder) SetConnectionString(connStr string) ISenderBuilder {
	if len(strings.TrimSpace(connStr)) > 0 {
		builder.connString = connStr
	}

	return builder
}

// SetOnAfterSendMessage(handler func(event *eventhub.Event)) If you wish to see which event was sent to event hubs, you can
// register to this handler.
func (builder *Builder) SetOnAfterSendMessage(handler func(event *eventhub.Event)) ISenderBuilder {
	if handler != nil {
		builder.onAfterSendMessage = handler
	}

	return builder
}

// SetOnBeforeSendMessage(handler func(event *eventhub.Event)) If you wish to see which event is about to be sent to event hubs, you can
// register to this handler.
func (builder *Builder) SetOnBeforeSendMessage(handler func(event *eventhub.Event)) ISenderBuilder {
	if handler != nil {
		builder.onBeforeSendMessage = handler
	}

	return builder
}

// SetOnAfterSendBatchMessage(handler func(batchSizeSent int, workerIndex int)) If you wish to see which amount of message was sent by batch
// you can register to this handle, it will delivery the amount of messages sent by a batch and with worker associated to that batch.
func (builder *Builder) SetOnAfterSendBatchMessage(handler func(batchSizeSent int, workerIndex int)) ISenderBuilder {
	if handler != nil {
		builder.onAfterSendBatchMessage = handler
	}

	return builder
}

// SetOnBeforeSendBatchMessage(handler func(batchSize int, workerIndex int)) If you wish to see which amount of batches is about to be sent
// you can register to this handle, it will delivery the amount of batches by worker associated to that batch. It batches can hold thousand of messages
// at once.
func (builder *Builder) SetOnBeforeSendBatchMessage(handler func(batchSize int, workerIndex int)) ISenderBuilder {
	if handler != nil {
		builder.onBeforeSendBatchMessage = handler
	}

	return builder
}

// GetSender() returns the mounted sender structure reference.
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

func (sender* Sender) SendEventsAsBatch(ctx context.Context, events *[]*eventhub.Event) error {
	limit, err := calcBatchLimitWithEvents(events)

	if err == nil {
		numGoRoutines := runtime.NumCPU()
		eventBatches := createEventBatchCollectionWithEvents(events, numGoRoutines, int64(limit), sender.numberOfMessages)
		if len(eventBatches) > 0 {
			var wg sync.WaitGroup

			sender.triggerBatches(ctx, &wg, numGoRoutines, eventBatches)
		}
	}

	return err
}

// SendMessage(message string, ctx context.Context) send a message to event hubs.
func (sender* Sender) SendMessage(message string, ctx context.Context) error {
	var err error = nil
	var i int64
	var mutex sync.Mutex

	for i = 0; i < sender.numberOfMessages; i++ {
		event := createAnEvent(sender.base64String, message, sender.messageSuffix)
		addProperties(event, sender.properties)

		if sender.onBeforeSendMessage != nil {
			mutex.Lock()
			sender.onBeforeSendMessage(event)
			mutex.Unlock()
		}

		runtime.Gosched()
		if err = sendMessage(sender.eHub, ctx, event); err != nil {
			break
		}

		if sender.onAfterSendMessage != nil {
			mutex.Lock()
			sender.onAfterSendMessage(event)
			mutex.Unlock()
		}
	}

	return err
}

// SendBatchMessage(message string, ctx context.Context) send a message to event hubs in batch.
// this function should be used together with SetNumberOfMessages and maybe SetMessageSuffix in the case you are not
// generating your own random content.
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

// AddProperties(properties map[string]interface{}) can be used to add properties using map format.
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
