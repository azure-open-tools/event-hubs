[![Build Status](https://github.com/azure-open-tools/event-hubs/workflows/Event-Hubs-Cli/badge.svg)

# Introduction
`Azure Event Hubs is a highly scalable publish-subscribe service that can ingest millions of 
events per second and stream them into multiple applications. This lets you process and analyze 
the massive amounts of data produced by your connected devices and applications. Once Event Hubs 
has collected the data, you can retrieve, transform and store it by using any real-time analytics 
provider or with batching/storage adapters.` - [Azure Event Hubs Official Client Library](https://github.com/Azure/azure-event-hubs-go/)

This library delivery an easier way of send and receive events from Azure Event Hubs. With this library
will be easy to create command line applications or transfer events over other protocol.

# Event Hub, event json structure

When you send a "message" to the event hub, you are sending an ```event``` and this ```event``` is a json like this:

```json
{
    "Data": "my message comes here, it can be a json{} or a byte array(string base64)",
    "PartitionKey": null,
    "ID": "MyId",
    "SystemProperties": {
        "SequenceNumber": 1,
        "EnqueuedTime": "2020-02-13T12:54:57.642Z",
        "Offset": 21479629240,
        "PartitionID": null,
        "PartitionKey": null
    }
}
```
Properties is an optional field, and you can also add them to enrich your event with metadata like this:
```json
{
    "Data": "my message comes here, can be a json{} or a byte array",
    "PartitionKey": null,
    "Properties": {
        "property1": "123",
        "property2": "metadata-2",
        "property3": "anything-else",
        ...
        ...
        ...
    },
    "ID": "MyId",
    "SystemProperties": {
        "SequenceNumber": 1,
        "EnqueuedTime": "2020-02-13T12:54:57.642Z",
        "Offset": 21479629240,
        "PartitionID": null,
        "PartitionKey": null
    }
}
```
If you didn't use Azure Events Hub yet, you can read more about it: [Event Hubs About](https://docs.microsoft.com/en-us/azure/event-hubs/event-hubs-about) and
[Event Hubs Features](https://docs.microsoft.com/en-us/azure/event-hubs/event-hubs-features).

# Sender usage (producer)

Go Modules:
```go
require github.com/azure-open-tools/event-hubs/sender v1.0.3
```
Import
```go
import github.com/azure-open-tools/event-hubs/sender
```

The connections string must be the 'SEND' action with the topic specified. 

Simple usage:
```go
builder := sender.NewSenderBuilder()
builder.SetConnectionString("Endpoint://....") //required field 
sender, err := builder.GetSender()

sender.SendMessage("myFirstMessage", context.Background())
```
Add properties/metadata to your event:
```go
builder := sender.NewSenderBuilder()
builder.SetConnectionString("Endpoint://....") //required field 
builder.AddProperties([]string{"propertyKey1:value1", "propertyKey2:value2", "<propertyKey>:<value>"}) //optional field
sender, err := builder.GetSender()

if err == nil {
    err = sender.SendMessage("myFirstMessage", context.Background())
}
```
Special fields:
* Base64 is an optional field. 
  if the message string you are passing to SendMessage function is a 
  byte array base64 encoded string, set this field to true.
  this way the library can convert back to byte array.
  this is necessary since the official event-hubs library encode to a base64 string
  before send.
```go
builder.SetBase64(true) //true or false. default value is FALSE
```
* NumberOfMessages is an optional field. you say to the library the amount of events you want to send at once.
it is good to use together with SendBatchMessage.
```go
builder.SetNumberOfMessages(int) //optional field. Here we can set how many messages(events) should be sent.
```
* Event Handlers. You can make the usage of event handlers to application usage.
```go
builder.SetOnAfterSendMessage(func (event *eventhub.Event) {})
builder.SetOnBeforeSendMessage(func (event *eventhub.Event) {})
```
* Batch Event Handlers, the send batch message process make usage of the amount of cpu available of the current system. It takes 
advantage of that to create as many go routines it cans to send events in parallel.
```go
builder.SetOnAfterSendBatchMessage(func (batchSize int, workerIndex int) {})
builder.SetOnBeforeSendBatchMessage(func (batchSizeSent int, workerIndex int){})
```

* Send batch events
```go
builder := sender.NewSenderBuilder()
builder.SetConnectionString("Endpoint://....") //required field 
builder.AddProperties([]string{"propertyKey1:value1", "propertyKey2:value2", "<propertyKey>:<value>"}) //optional field
builder.SetNumberOfMessages(1000000) //1mi events
sender, err := builder.GetSender()

if err == nil {
    err = sender.SendBatchMessage(message, context.Background())
}

```
# Receiver usage (consumer)
Go Modules:
```go
require github.com/azure-open-tools/event-hubs/receiver v1.0.1
```
Import
```go
import github.com/azure-open-tools/event-hubs/receiver
```

The connections string must be the 'LISTEN' action with the topic specified.

Simple usage:
```go
builder := receiver.NewReceiverBuilder()

if builder != nil {
    builder.SetConnectionString("Endpoint://") //required field
    builder.SetReceiverHandler(OnReceiverHandler)

    rcv, err := builder.GetReceiver()
    if err == nil {
        err = rcv.StartListener(context.Background())
    }
}
```
* recommended usage:
Consumer group is an optional field, but recommended (you must create it first in azure portal or azure cli. 
$Default consumer group will be used as default value. if you are using against the development environment it's not a problem.
However against Production environment others clients could be disconnected it depends on your Azure Event Hubs settings.
```go
builder := receiver.NewReceiverBuilder()

if builder != nil {
    builder.SetConnectionString("Endpoint://") //required field
    builder.SetConsumerGroup("debug") //recommended
    builder.SetReceiverHandler(func(ctx context.Context, event *eventhub.Event) error { })

    rcv, err := builder.GetReceiver()
    if err == nil {
        err = rcv.StartListener(context.Background())
    }
}
```
* ReceiverHandler it is the event handler you must add to handle the events arrive from the event hubs.
Whenever an event arrives this event handler will be executed.

```go
builder.SetReceiverHandler(func(ctx context.Context, event *eventhub.Event) error {
    fmt.Println(event.Data)
})
```
* Special fields:
    * DataFilers(event.Data field): you can set any kind of string, if the string match the event will be delivered to the ReceiverHandler function.
    * PropertyFilters(event.Properties field): same as above, but will concentrate into the property fields.
    * ListenerPartitionIds: here you can specify which partition ids you want to listen to.
    if you let it away, the library will listen all partitionIds available.
```go
builder := receiver.NewReceiverBuilder()

if builder != nil {
    builder.SetConnectionString("Endpoint://") //required field
    builder.SetConsumerGroup("debug") //recommended
    builder.AddDataFilters("any content to be filtered")
    builder.AddPropertyFilters([]string{"propertyKey1:value1", "propertyKey2:value2"})
    builder.AddListenerPartitionIds([]string{"0", "1", "12", "21"}) 
    builder.SetReceiverHandler(func(ctx context.Context, event *eventhub.Event) error {
                                   fmt.Println(event.Data)
                               })

    rcv, err := builder.GetReceiver()
    if err == nil {
        err = rcv.StartListener(context.Background())
    }
}
```
