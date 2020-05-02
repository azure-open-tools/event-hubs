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