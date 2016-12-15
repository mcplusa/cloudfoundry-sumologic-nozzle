package sumoCFFirehose

import (
	"bytes"
	"fmt"
	"net/http"
	"time"

	"bitbucket.org/mcplusa-ondemand/firehose-to-sumologic/eventQueue"
)

type SumoLogicAppender struct {
	url                      string
	connectionTimeout        int //10000
	httpClient               http.Client
	nozzleQueue              *eventQueue.Queue
	eventsBatchSize          int
	logEventsInCurrentBuffer int
	logStringToSend          *bytes.Buffer
}

func NewSumoLogicAppender(urlValue string, connectionTimeoutValue int, nozzleQueue *eventQueue.Queue, eventsBatchSize int) *SumoLogicAppender {
	return &SumoLogicAppender{
		url:                      urlValue,
		connectionTimeout:        connectionTimeoutValue,
		httpClient:               http.Client{Timeout: time.Duration(connectionTimeoutValue * int(time.Millisecond))},
		nozzleQueue:              nozzleQueue,
		eventsBatchSize:          eventsBatchSize,
		logEventsInCurrentBuffer: 0,
		logStringToSend:          bytes.NewBufferString(""),
	}
}

func (s *SumoLogicAppender) Start() {
	timer := time.Now()
	fmt.Println("Starting Appender Worker")
	for {
		time.Sleep(300 * time.Millisecond)
		if s.nozzleQueue.GetCount() != 0 { //if queue is not empty, AppendLogs
			s.AppendLogs()
		}
		if s.logEventsInCurrentBuffer >= s.eventsBatchSize { //if buffer is full, send logs to sumo
			fmt.Println("Buffer full, sending logs to sumo...")

			s.SendToSumo(s.logStringToSend)

		} else if time.Since(timer).Seconds() >= 10 { // else if timer is up, send existing logs to sumo
			fmt.Println("timer finished, sending logs...")
			s.SendToSumo(s.logStringToSend)
			timer = time.Now() //reset timer
		}

	}
}

func StringBuilder(node *eventQueue.Node) string {
	buf := new(bytes.Buffer)
	if node.Event.Fields["message_type"] == nil {
		return ""
	}
	if node.Event.Fields["message_type"] == "" {
		return ""
	}
	message := time.Unix(0, node.Event.Fields["timestamp"].(int64)*int64(time.Nanosecond)).String() + "\t" + node.Event.Fields["message_type"].(string) + "\t" + node.Event.Msg + "\n"
	buf.WriteString(message)

	return buf.String()
}

func (s *SumoLogicAppender) AppendLogs() {
	// the appender calls for the next message in the queue and parse it to a string
	//timer := time.NewTimer(60 * time.Second)
	s.logStringToSend.Write([]byte(StringBuilder(s.nozzleQueue.Pop())))
	s.logEventsInCurrentBuffer++
}

func (s *SumoLogicAppender) SendToSumo(log *bytes.Buffer) {
	fmt.Println(log)
	request, err := http.NewRequest("POST", s.url, log)
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
	s.logEventsInCurrentBuffer = 0                // reset counter
	s.logStringToSend = bytes.NewBufferString("") //reset String
	defer response.Body.Close()

}
