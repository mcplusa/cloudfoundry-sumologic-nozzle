package sumoCFFirehose

import (
	"testing"

	"github.com/stretchr/testify/assert"

	. "bitbucket.org/mcplusa-ondemand/firehose-to-sumologic/eventQueue"
	. "bitbucket.org/mcplusa-ondemand/firehose-to-sumologic/events"
)

func TestAppenderStringBuilder(t *testing.T) {
	event1 := Event{
		Fields: map[string]interface{}{
			"deployment": "cf",
			"ip":         "10.193.166.33",
			"job":        "cloud_controller",
			"job_index":  "c82feee9-2159-4b05-b669-a9929eb59017",
			"name":       "requests.completed",
			"origin":     "cc",
			"unit":       "counter",
			"value":      558108,
		},
		Msg:  "",
		Type: "ValueMetric",
	}

	event2 := Event{
		Fields: map[string]interface{}{
			"delta":      9,
			"deployment": "cf-redis",
			"ip":         "10.193.166.84",
			"job":        "dedicated-node",
			"job_index":  "8081eca4-9e27-49cb-83ce-948e703c0939",
			"name":       "dropsondeMarshaller.sentEnvelopes",
			"origin":     "MetronAgent",
			"total":      10249446,
		},
		Msg:  "",
		Type: "CounterEvent",
	}

	event3 := Event{
		Fields: map[string]interface{}{
			"delta":      582,
			"deployment": "cf-redis",
			"ip":         "10.193.166.84",
			"job":        "dedicated-node",
			"job_index":  "23f9be01-bd83-4967-acba-69fc649f4ee6",
			"name":       "dropsondeAgentListener.receivedByteCount",
			"origin":     "MetronAgent",
			"total":      639557085,
		},
		Msg:  "",
		Type: "CounterEvent",
	}
	queue := Queue{
		Events: make([]*Event, 3),
	}
	queue.Push(&event1)
	queue.Push(&event2)
	queue.Push(&event3)

	finalString := ""
	for queue.GetCount() > 0 {
		finalString = finalString + StringBuilder(queue.Pop(), true)
	}
	assert.Equal(t, finalString, "{\"Fields\":{\"deployment\":\"cf\",\"ip\":\"10.193.166.33\",\"job\":\"cloud_controller\",\"job_index\":\"c82feee9-2159-4b05-b669-a9929eb59017\",\"name\":\"requests.completed\",\"origin\":\"cc\",\"unit\":\"counter\",\"value\":558108},\"Msg\":\"\",\"Type\":\"ValueMetric\"}\n"+
		"{\"Fields\":{\"delta\":9,\"deployment\":\"cf-redis\",\"ip\":\"10.193.166.84\",\"job\":\"dedicated-node\",\"job_index\":\"8081eca4-9e27-49cb-83ce-948e703c0939\",\"name\":\"dropsondeMarshaller.sentEnvelopes\",\"origin\":\"MetronAgent\",\"total\":10249446},\"Msg\":\"\",\"Type\":\"CounterEvent\"}\n"+
		"{\"Fields\":{\"delta\":582,\"deployment\":\"cf-redis\",\"ip\":\"10.193.166.84\",\"job\":\"dedicated-node\",\"job_index\":\"23f9be01-bd83-4967-acba-69fc649f4ee6\",\"name\":\"dropsondeAgentListener.receivedByteCount\",\"origin\":\"MetronAgent\",\"total\":639557085},\"Msg\":\"\",\"Type\":\"CounterEvent\"}\n", "")
}

func TestStringBuilderVerboseLogsFalse(t *testing.T) {
	eventVerboseLogMessage := Event{
		Fields: map[string]interface{}{
			"message_type":    "OUT",
			"source_instance": 0,
			"deployment":      "cf",
			"ip":              "10.193.166.47",
			"job":             "diego_cell",
			"job_index":       "c62aebe5-16b8-43f5-a589-1267e09b9537",
			"cf_ignored_app":  "false",
			"timestamp":       int64(1483629662001580713),
			"source_type":     "APP",
			"origin":          "rep",
			"cf_app_id":       "7833dc75-4484-409c-9b74-90b6454906c6",
		},
		Msg:  "Triggering 'app usage events fetcher'",
		Type: "LogMessage",
	}

	finalMessage := StringBuilder(&eventVerboseLogMessage, false)
	assert.NotContains(t, finalMessage, "source_type", "dsds")

}

func TestStringBuilderVerboseLogsTrue(t *testing.T) {
	eventVerboseLogMessage := Event{
		Fields: map[string]interface{}{
			"message_type":    "OUT",
			"source_instance": 0,
			"deployment":      "cf",
			"ip":              "10.193.166.47",
			"job":             "diego_cell",
			"job_index":       "c62aebe5-16b8-43f5-a589-1267e09b9537",
			"cf_ignored_app":  "false",
			"timestamp":       int64(1483629662001580713),
			"source_type":     "APP",
			"origin":          "rep",
			"cf_app_id":       "7833dc75-4484-409c-9b74-90b6454906c6",
		},
		Msg:  "Triggering 'app usage events fetcher'",
		Type: "LogMessage",
	}
	finalMessage := StringBuilder(&eventVerboseLogMessage, true)

	assert.Contains(t, finalMessage, "source_type", "")

}
func TestSendParseCustomMetadata(t *testing.T) {
	customMetadata := "Key1:Value1,Key2:Value2,Key3:Value3"
	mapCustomMetadata := ParseCustomMetadata(customMetadata)
	mapExpected := map[string]string{
		"Key1": "Value1",
		"Key2": "Value2",
		"Key3": "Value3",
	}

	assert.Equal(t, mapExpected, mapCustomMetadata, "")

}
