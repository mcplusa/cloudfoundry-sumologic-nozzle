package eventQueue

import (
	"testing"

	. "bitbucket.org/mcplusa-ondemand/firehose-to-sumologic/events"
	"github.com/stretchr/testify/assert"
)

func TestQueueFIFO(t *testing.T) {

	node1 := Node{
		event: Event{
			Fields: map[string]interface{}{
				"message_type": "OUT",
				"cf_app_id":    "011",
			},
			Msg: "index [01]",
		},
	}
	node2 := Node{
		event: Event{
			Fields: map[string]interface{}{
				"message_type": "OUT",
				"cf_app_id":    "022",
			},
			Msg: "index [02]",
		},
	}
	node3 := Node{
		event: Event{
			Fields: map[string]interface{}{
				"message_type": "OUT",
				"cf_app_id":    "033",
			},
			Msg: "index [03]",
		},
	}

	queue := Queue{
		nodes: make([]*Node, 3),
	}

	queue.Push(&node1)
	queue.Push(&node2)
	queue.Push(&node3)

	assert.Equal(t, queue.Pop().event.Msg, "index [01]", "")
	assert.Equal(t, queue.Pop().event.Msg, "index [02]", "")
	assert.Equal(t, queue.Pop().event.Msg, "index [03]", "")
}
