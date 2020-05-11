package sender

import (
	eventhub "github.com/Azure/azure-event-hubs-go/v3"
	"strings"
	"testing"
)

func TestSenderBuilder_AddProperty(t *testing.T) {
	builder := NewSenderBuilder()

	if builder != nil {
		builder.AddProperty("property1:value1")
		if len(builder.properties) != 1 {
			t.Error("collection should have 1 element")
		}
	} else {
		t.Error("builder not instantiated")
	}
}

func TestSenderBuilder_AddProperties(t *testing.T) {
	builder := NewSenderBuilder()

	if builder != nil {
		builder.AddProperties([]string{"property1:value1", "property2:value2", "property3:value3"})
		if len(builder.properties) != 3 {
			t.Error("collection should have 3 element")
		}
	} else {
		t.Error("builder not instantiated")
	}
}

func TestSender_AddProperties_To_Event(t *testing.T) {
	var properties = []string{"iothub-enqueuedtime:1580737059326;serviceid:11;functionid:15;partnerId1:9b1f9625-09e1-4d1f-a0a0-c0d50b001c5d;trackingid:ctp-1234coba01-integration"}
	var event = &eventhub.Event{}
	addProperties(event, properties)

	if len(event.Properties) == 0 {
		t.Error("properties was not added successfully")
	}
}

func TestSender_AddProperty_To_Event(t *testing.T) {
	var properties = []string{"iothub-enqueuedtime:1580737059326"}
	var event = &eventhub.Event{}
	addProperties(event, properties)

	if len(event.Properties) != 1 {
		t.Error("properties was not added successfully")
	}
}

func TestSenderBuilder_AddPartitionId(t *testing.T) {
	builder := NewSenderBuilder()

	if builder != nil {
		builder.AddPartitionId("0")
		if len(builder.partitionIds) != 1 {
			t.Error("collection should have 1 element")
		}
	} else {
		t.Error("builder not instantiated")
	}
}

func TestSenderBuilder_AddPartitionIds(t *testing.T) {
	builder := NewSenderBuilder()

	if builder != nil {
		builder.AddPartitionIds([]string{"0", "1", "2"})
		if len(builder.partitionIds) != 3 {
			t.Error("collection should have 3 element")
		}
	} else {
		t.Error("builder not instantiated")
	}
}

func TestSenderBuilder_SetConnectionString(t *testing.T) {
	builder := NewSenderBuilder()

	if builder != nil {
		builder.SetConnectionString("endpoint://...")
		if len(strings.TrimSpace(builder.connString)) == 0 {
			t.Errorf("string should not be empty")
		}
	} else {
		t.Error("builder not instantiated")
	}
}

func TestSenderBuilder_SetBase64(t *testing.T) {
	builder := NewSenderBuilder()

	if builder != nil {
		builder.SetBase64(true)
		if builder.base64String == false {
			t.Errorf("should be true")
		}
	} else {
		t.Error("builder not instantiated")
	}
}

func TestSenderBuilder_GetSender(t *testing.T) {
	builder := NewSenderBuilder()

	if builder != nil {
		receiver, err := builder.GetSender()
		if receiver != nil && err == nil {
			t.Error("sender should be nil and error different nil, conn string is mandatory field")
		}
	} else {
		t.Error("builder not instantiated")
	}
}

func TestSenderBuilder_GetSender_Regular_Usage(t *testing.T) {
	builder := NewSenderBuilder()

	if builder != nil {
		builder.SetConnectionString("endpoint://...")
		builder.AddProperty("property1:value1")
		builder.AddPartitionId("0")
		builder.SetBase64(true)

		sender, _ := builder.GetSender()

		if sender != nil {
			if len(sender.properties) != 1 {
				t.Error("collection should have 1 element")
			}
			if len(sender.partitionIds) != 1 {
				t.Error("collection should have 1 element")
			}
			if len(strings.TrimSpace(sender.connString)) == 0 {
				t.Error("string should not be empty")
			}
			if sender.base64String == false {
				t.Error("base64 should be true")
			}
		} else {
			t.Error("sender should not be null")
		}
	} else {
		t.Error("builder not instantiated")
	}
}