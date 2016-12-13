package sumoCFFirehose

import (
	"bytes"
	"fmt"
	"net"
	"net/http"
	"time"

	"bitbucket.org/mcplusa-ondemand/firehose-to-sumologic/eventQueue"
)

type SumoLogicAppender struct {
	url               string
	connectionTimeout int //10000
	httpClient        http.Client
	nozzleQueue       eventQueue.Queue
	eventsBatch       int
}

func NewSumoLogicAppender(urlValue string, connectionTimeoutValue int, nozzleQueue eventQueue.Queue, eventsBatch int) *SumoLogicAppender {
	return &SumoLogicAppender{
		url:               urlValue,
		connectionTimeout: connectionTimeoutValue,
		httpClient:        http.Client{Timeout: time.Duration(connectionTimeoutValue * int(time.Millisecond))},
		nozzleQueue:       nozzleQueue,
		eventsBatch:       eventsBatch,
	}
}

func (s *SumoLogicAppender) Connect() bool {
	success := false
	if s.url != "" {
		conn, err := net.Dial("tcp", s.url)
		if err != nil {
			fmt.Printf(fmt.Sprintf("Unable to connect to sumo server [%s]!\n", s.url), err.Error())
		} else {

			fmt.Printf(fmt.Sprintf("Connected to [%s]!\n", s.url), false)
			success = true
			defer conn.Close()
		}
	}

	return success
}

func StringBuilder(queue eventQueue.Queue) string {
	buf := new(bytes.Buffer)
	for queue.GetCount() > 0 { //Pop eventsBatch from queue
		event := queue.Pop().GetNodeEvent()
		if event.Fields["message_type"] == nil {
			return ""
		}
		if event.Fields["message_type"] == "" {
			return ""
		}
		message := time.Unix(0, event.Fields["timestamp"].(int64)*int64(time.Nanosecond)).String() + "\t" + event.Fields["message_type"].(string) + "\t" + event.Msg + "\n"
		buf.WriteString(message)
	}

	return buf.String()
}

func (s *SumoLogicAppender) AppendLogs(queue eventQueue.Queue) {
	// the appender calls for the next message in the queue and parse it to a string
	if queue.GetCount() == s.eventsBatch { //whent the batch limit is met, call stringBuilder
		logMessage := StringBuilder(queue)
		s.SendToSumo(logMessage)
	}

}

func (s *SumoLogicAppender) SendToSumo(log string) {
	request, err := http.NewRequest("POST", s.url, bytes.NewBufferString(log))
	if err != nil {
		fmt.Printf("http.NewRequest() error: %v\n", err)
		return
	}
	//request.Header.Add("content-type", "application/json")
	//request.SetBasicAuth("admin", "admin")
	response, err := s.httpClient.Do(request)

	if err != nil {
		fmt.Printf("http.Do() error: %v\n", err)
		return
	} else {
		fmt.Println("Do(Request) successful")
	}
	defer response.Body.Close()

}
