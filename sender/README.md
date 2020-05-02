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