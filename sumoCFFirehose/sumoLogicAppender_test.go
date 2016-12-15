package sumoCFFirehose

import (
	"testing"

	"github.com/stretchr/testify/assert"

	. "bitbucket.org/mcplusa-ondemand/firehose-to-sumologic/eventQueue"
	. "bitbucket.org/mcplusa-ondemand/firehose-to-sumologic/events"
)

func testAppenderStringBuilder(t *testing.T) {
	event1 := Event{
		Fields: map[string]interface{}{
			"timestamp":    "1481569361828366387",
			"message_type": "OUT",
			"cf_app_id":    "011",
		},
		Msg: "index [01]",
	}

	event2 := Event{
		Fields: map[string]interface{}{
			"timestamp":    "1481569362844737993",
			"message_type": "OUT",
			"cf_app_id":    "022",
		},
		Msg: "index [02]",
	}

	event3 := Event{
		Fields: map[string]interface{}{
			"timestamp":    "1481569363862436654",
			"message_type": "OUT",
			"cf_app_id":    "033",
		},
		Msg: "index [03]",
	}

	queue := Queue{
		Events: make([]*Event, 3),
	}
	queue.Push(&event1)
	queue.Push(&event2)
	queue.Push(&event3)

	finalString := ""
	for queue.GetCount() > 0 {
		finalString = finalString + StringBuilder(queue.Pop())
	}
	assert.Equal(t, finalString, "2016-12-12 16:02:41.828366387 -0300 CLST"+"\t"+"OUT"+"\t"+"index [01]"+"\n"+
		"2016-12-12 16:02:42.844737993 -0300 CLST"+"\t"+"OUT"+"\t"+"index [02]"+"\n"+
		"2016-12-12 16:02:43.862436654 -0300 CLST"+"\t"+"OUT"+"\t"+"index [03]"+"\n", "")
}
