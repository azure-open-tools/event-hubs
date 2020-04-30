package sender

import (
	"runtime"
	"testing"
)

func TestEventBatch_with_100_messages_with_limit_10_return_with_success(t *testing.T) {
	message := "test-message"
	sender := &Sender{}
	sender.base64String = false
	sender.properties = []string{"messageId:1234" }

	eventBatches := createEventBatchCollection(sender, runtime.NumCPU(), 10, 100, message, false)

	if len(eventBatches) == 9 {
		for i := 0; i < len(eventBatches); i++ {
			events, _ := eventBatches[i].Get(0)
			if i == 8 && len(events) == 2 {
				if eventBatches[i].Size() == 10  {
					break
				}
			}
			if len(events) != 10 {
				t.Errorf("batch with the index %v have %v but should be 10", i, len(events))
				break
			}
		}
	} else {
		t.Error("event batches is greater or minor than 10")
	}
}

func TestEventBatch_with_100_messages_and_real_limit_return_one_batch_with_100_messages_success(t *testing.T) {
	message := "test-message"
	sender := &Sender{}
	sender.base64String = false
	sender.properties = []string{"messageId:1234" }

	limit, _ := calcBatchLimit(sender, message, false)
	eventBatches := createEventBatchCollection(sender, runtime.NumCPU(), int64(limit), 100, message, false)

	if len(eventBatches) == 1 {
		events, _ := eventBatches[0].Get(0)
		if len(events) != 100 {
			t.Errorf("batch with the index %v have %v but should be 100", 0, eventBatches[0].Size())
		}
	} else {
		t.Error("Test is broken")
	}
}

func TestEventBatch_with_1million_messages_and_real_limit_success(t *testing.T) {
	message := "test-message"
	sender := &Sender{}
	sender.base64String = false
	sender.properties = []string{"messageId:1234" }

	limit, _ := calcBatchLimit(sender, message, false)
	eventBatches := createEventBatchCollection(sender, runtime.NumCPU(), int64(limit), 1000000, message, false)

	if len(eventBatches) == runtime.NumCPU() {
		var count int
		for i := 0; i < len(eventBatches); i++ {
			for j := 0; j < eventBatches[i].Size(); j++ {
				array, _ := eventBatches[i].Get(j)
				count = count + len(array)
				//fmt.Println(len(array))
			}
		}
		if count < 1000000 || count > 1000000 {
			t.Error("count is less or greater than 1000000")
		}
	} else {
		t.Error("Test is broken")
	}
}