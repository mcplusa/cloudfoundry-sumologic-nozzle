package sumoCFFirehose

import (
	"testing"

	. "bitbucket.org/mcplusa-ondemand/firehose-to-sumologic/eventQueue"
	. "bitbucket.org/mcplusa-ondemand/firehose-to-sumologic/events"
	"github.com/stretchr/testify/assert"
)

func testAppenderStringBuilder(t *testing.T) {

	node1 := Node{
		Event: Event{
			Fields: map[string]interface{}{
				"timestamp":    "1481569361828366387",
				"message_type": "OUT",
				"cf_app_id":    "011",
			},
			Msg: "index [01]",
		},
	}
	node2 := Node{
		Event: Event{
			Fields: map[string]interface{}{
				"timestamp":    "1481569362844737993",
				"message_type": "OUT",
				"cf_app_id":    "022",
			},
			Msg: "index [02]",
		},
	}
	node3 := Node{
		Event: Event{
			Fields: map[string]interface{}{
				"timestamp":    "1481569363862436654",
				"message_type": "OUT",
				"cf_app_id":    "033",
			},
			Msg: "index [03]",
		},
	}

	queue := Queue{
		Nodes: make([]*Node, 3),
	}
	queue.Push(&node1)
	queue.Push(&node2)
	queue.Push(&node3)

	finalString := ""
	for queue.GetCount() > 0 {
		finalString = finalString + StringBuilder(queue.Pop())
	}
	assert.Equal(t, finalString, "2016-12-12 16:02:41.828366387 -0300 CLST"+"\t"+"OUT"+"\t"+"index [01]"+"\n"+
		"2016-12-12 16:02:42.844737993 -0300 CLST"+"\t"+"OUT"+"\t"+"index [02]"+"\n"+
		"2016-12-12 16:02:43.862436654 -0300 CLST"+"\t"+"OUT"+"\t"+"index [03]"+"\n", "")

}
