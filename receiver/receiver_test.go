package receiver

import (
	"context"
	eventhub "github.com/Azure/azure-event-hubs-go/v3"
	"strings"
	"testing"
)

func TestReceiverBuilder_AddDataFilter(t *testing.T) {
	builder := NewReceiverBuilder()

	if builder != nil {
		builder.AddDataFilter("filter1")
		if len(builder.DataFilter) != 1 {
			t.Error("collection should have 1 element")
		}
	} else {
		t.Error("builder not instantiated")
	}
}

func TestReceiverBuilder_AddDataFilters(t *testing.T) {
	builder := NewReceiverBuilder()

	if builder != nil {
		builder.AddDataFilters([]string{"filter1", "filter2", "filter3"})
		if len(builder.DataFilter) != 3 {
			t.Error("collection should have 3 element")
		}
	} else {
		t.Error("builder not instantiated")
	}
}

func TestReceiverBuilder_AddPropertyFilter(t *testing.T) {
	builder := NewReceiverBuilder()

	if builder != nil {
		builder.AddPropertyFilter("filter1")
		if len(builder.PropertyFilter) != 1 {
			t.Error("collection should have 1 element")
		}
	} else {
		t.Error("builder not instantiated")
	}
}

func TestReceiverBuilder_AddPropertyFilters(t *testing.T) {
	builder := NewReceiverBuilder()

	if builder != nil {
		builder.AddPropertyFilters([]string{"filter1", "filter2", "filter3"})
		if len(builder.PropertyFilter) != 3 {
			t.Error("collection should have 3 element")
		}
	} else {
		t.Error("builder not instantiated")
	}
}

func TestReceiverBuilder_AddListenerPartitionId(t *testing.T) {
	builder := NewReceiverBuilder()

	if builder != nil {
		builder.AddListenerPartitionId("0")
		if len(builder.PartitionIds) != 1 {
			t.Error("collection should have 1 element")
		}
	} else {
		t.Error("builder not instantiated")
	}
}

func TestReceiverBuilder_AddListenerPartitionIds(t *testing.T) {
	builder := NewReceiverBuilder()

	if builder != nil {
		builder.AddListenerPartitionIds([]string{"0", "1", "2"})
		if len(builder.PartitionIds) != 3 {
			t.Error("collection should have 3 element")
		}
	} else {
		t.Error("builder not instantiated")
	}
}

func TestReceiverBuilder_SetConnectionString(t *testing.T) {
	builder := NewReceiverBuilder()

	if builder != nil {
		builder.SetConnectionString("endpoint://...")
		if len(strings.TrimSpace(builder.ConnString)) == 0 {
			t.Errorf("string should not be empty")
		}
	} else {
		t.Error("builder not instantiated")
	}
}

func TestReceiverBuilder_SetConsumerGroup(t *testing.T) {
	builder := NewReceiverBuilder()

	if builder != nil {
		builder.SetConsumerGroup("AnyString")
		if len(strings.TrimSpace(builder.ConsumerGroup)) == 0 {
			t.Errorf("string should not be empty")
		}
	} else {
		t.Error("builder not instantiated")
	}
}

func TestReceiverBuilder_GetReceiver(t *testing.T) {
	builder := NewReceiverBuilder()

	if builder != nil {
		receiver, err := builder.GetReceiver()
		if receiver != nil && err == nil {
			t.Error("receiver should be nil and error different nil, conn string is mandatory field")
		}
	} else {
		t.Error("builder not instantiated")
	}
}

func TestReceiverBuilder_GetReceiver_With_Default_ConsumerGroup(t *testing.T) {
	builder := NewReceiverBuilder()

	if builder != nil {
		builder.SetConnectionString("endpoint://...")
		receiver, err := builder.GetReceiver()
		if receiver == nil && err != nil {
			t.Error("receiver should an instance and error should be nil")
		}

		if receiver != nil && receiver.consumerGroup != "$Default" {
			t.Error("consumer group should be equals to $Default string")
		}
	} else {
		t.Error("builder not instantiated")
	}
}

func TestReceiverBuilder_GetReceiver_Without_Default_ConsumerGroup(t *testing.T) {
	builder := NewReceiverBuilder()

	if builder != nil {
		builder.SetConnectionString("endpoint://...")
		builder.SetConsumerGroup("monitor")
		receiver, err := builder.GetReceiver()
		if receiver == nil && err != nil {
			t.Error("receiver should an instance and error should be nil")
		}

		if receiver != nil && receiver.consumerGroup != "monitor" {
			t.Error("consumer group should be equals to monitor string")
		}
	} else {
		t.Error("builder not instantiated")
	}
}

func TestReceiverBuilder_GetReceiver_Regular_Usage(t *testing.T) {
	builder := NewReceiverBuilder()

	if builder != nil {
		builder.SetConnectionString("endpoint://...")
		builder.SetConsumerGroup("monitor")
		builder.AddDataFilter("filter1")
		builder.AddPropertyFilter("filter2")
		builder.AddListenerPartitionId("0")
		builder.SetReceiverHandler(func(ctx context.Context, event *eventhub.Event) error {
			return nil
		})

		receiver, _ := builder.GetReceiver()

		if receiver != nil {
			if len(receiver.dataFilter) != 1 {
				t.Error("collection should have 1 element")
			}
			if len(receiver.propertyFilter) != 1 {
				t.Error("collection should have 1 element")
			}
			if len(receiver.partitionIds) != 1 {
				t.Error("collection should have 1 element")
			}
			if len(strings.TrimSpace(receiver.connString)) == 0 {
				t.Error("string should not be empty")
			}
			if len(strings.TrimSpace(receiver.consumerGroup)) == 0 {
				t.Error("string should not be empty")
			}
			if receiver.consumerGroup != "monitor" {
				t.Error("consumer group should be equals to monitor string")
			}
			if receiver.onReceiveHandler == nil {
				t.Error("event handler should be different from nil")
			}
		} else {
			t.Error("receiver should not be null")
		}
	} else {
		t.Error("builder not instantiated")
	}
}

func TestReceiver_Event_Handler(t *testing.T) {
	var result = false
	builder := NewReceiverBuilder()

	if builder != nil {
		builder.SetConnectionString("endpoint://...")
		builder.SetReceiverHandler(func(ctx context.Context, event *eventhub.Event) error {
			result = true
			return nil
		})

		receiver, _ := builder.GetReceiver()
		err := receiver.onReceive(context.Background(), &eventhub.Event{})
		if result == false && err != nil {
			t.Error("event should arrive to the caller")
		}
	} else {
		t.Error("builder not instantiated")
	}
}

func TestReceiver_Event_Handler_With_Filter_Applied_Against_Different_Value_Expected_Event_NotTriggered(t *testing.T) {
	var result = false
	builder := NewReceiverBuilder()

	if builder != nil {
		builder.SetConnectionString("endpoint://...")
		builder.SetConsumerGroup("monitor")
		builder.AddDataFilter("filter1")
		builder.SetReceiverHandler(func(ctx context.Context, event *eventhub.Event) error {
			result = true
			return nil
		})

		receiver, _ := builder.GetReceiver()
		err := receiver.onReceive(context.Background(), &eventhub.Event{})
		if result == true && err != nil {
			t.Error("event should arrive to the caller")
		}
	} else {
		t.Error("builder not instantiated")
	}
}

func TestReceiver_Event_Handler_With_Filter_Applied_Against_Correct_Value_Expected_Event_Triggered(t *testing.T) {
	var result = false
	builder := NewReceiverBuilder()

	if builder != nil {
		builder.SetConnectionString("endpoint://...")
		builder.SetConsumerGroup("monitor")
		builder.AddDataFilter("filter1")
		builder.SetReceiverHandler(func(ctx context.Context, event *eventhub.Event) error {
			result = true
			return nil
		})

		receiver, _ := builder.GetReceiver()
		err := receiver.onReceive(context.Background(), &eventhub.Event{ Data: []byte("test filter1") })
		if result == false && err != nil {
			t.Error("event should arrive to the caller")
		}
	} else {
		t.Error("builder not instantiated")
	}
}

func TestReceiver_Event_Handler_With_PropertyFilter_Applied_Against_Correct_Value_Expected_Event_Triggered(t *testing.T) {
	var result = false
	var properties = make(map[string]interface{})
	properties["messageId"] = "value1"

	builder := NewReceiverBuilder()

	if builder != nil {
		builder.SetConnectionString("endpoint://...")
		builder.SetConsumerGroup("monitor")
		builder.AddDataFilter("filter1")
		builder.AddPropertyFilter("value1")
		builder.SetReceiverHandler(func(ctx context.Context, event *eventhub.Event) error {
			result = true
			return nil
		})

		receiver, _ := builder.GetReceiver()
		err := receiver.onReceive(context.Background(), &eventhub.Event {
			Data: []byte("test filter2"),
			Properties: properties,
		})
		if result == false && err != nil {
			t.Error("event should arrive to the caller")
		}
	} else {
		t.Error("builder not instantiated")
	}
}

func TestReceiver_Event_Handler_With_PropertyFilter_SpecificProperty_Applied_Against_Correct_Value_Expected_Event_Triggered(t *testing.T) {
	var result = false
	var properties = make(map[string]interface{})
	properties["messageId"] = "value1"

	builder := NewReceiverBuilder()

	if builder != nil {
		builder.SetConnectionString("endpoint://...")
		builder.SetConsumerGroup("monitor")
		builder.AddDataFilter("filter1")
		builder.AddPropertyFilter("messageId:value1")
		builder.SetReceiverHandler(func(ctx context.Context, event *eventhub.Event) error {
			result = true
			return nil
		})

		receiver, _ := builder.GetReceiver()
		err := receiver.onReceive(context.Background(), &eventhub.Event {
			Data: []byte("test filter2"),
			Properties: properties,
		})
		if result == false && err != nil {
			t.Error("event should arrive to the caller")
		}
	} else {
		t.Error("builder not instantiated")
	}
}

func TestReceiver_Event_Handler_With_PropertyFilter_SpecificProperty_Applied_Against_Wrong_Value_Expected_Event_NotTriggered(t *testing.T) {
	var result = false
	var properties = make(map[string]interface{})
	properties["messageId"] = "value1"

	builder := NewReceiverBuilder()

	if builder != nil {
		builder.SetConnectionString("endpoint://...")
		builder.SetConsumerGroup("monitor")
		builder.AddDataFilter("filter1")
		builder.AddPropertyFilter("messageId:value2")
		builder.SetReceiverHandler(func(ctx context.Context, event *eventhub.Event) error {
			result = true
			return nil
		})

		receiver, _ := builder.GetReceiver()
		err := receiver.onReceive(context.Background(), &eventhub.Event {
			Data: []byte("test filter2"),
			Properties: properties,
		})
		if result == false && err != nil {
			t.Error("event should arrive to the caller")
		}
	} else {
		t.Error("builder not instantiated")
	}
}