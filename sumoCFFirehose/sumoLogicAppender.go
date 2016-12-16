package sumoCFFirehose

import (
	"bytes"
	"net/http"
	"time"

	"bitbucket.org/mcplusa-ondemand/firehose-to-sumologic/eventQueue"
	"bitbucket.org/mcplusa-ondemand/firehose-to-sumologic/events"
	"bitbucket.org/mcplusa-ondemand/firehose-to-sumologic/logging"
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
	logging.Info.Println("Starting Appender Worker")
	for {
		time.Sleep(300 * time.Millisecond)
		// while queue is not empty && s.eventsBatchSize not completed, queue.POP (appendLogs)
		for s.nozzleQueue.GetCount() != 0 && s.logEventsInCurrentBuffer <= s.eventsBatchSize {
			s.AppendLogs()                                       //this method POP an event from queue
			timer = time.Now()                                   //reset timer
			if s.logEventsInCurrentBuffer == s.eventsBatchSize { //if buffer is full, send logs to sumo
				logging.Trace.Println("Batch Size complete")
				s.SendToSumo(s.logStringToSend)
				break
			} else if time.Since(timer).Seconds() >= 10 { // else if timer is up, send existing logs to sumo
				logging.Trace.Println("Sending current batch of logs after timer exceeded limit")
				s.SendToSumo(s.logStringToSend)
				timer = time.Now() //reset timer
				break
			}
		}
	}
}

func StringBuilder(event *events.Event) string {
	buf := new(bytes.Buffer)
	if event.Fields["message_type"] == nil {
		return ""
	}
	if event.Fields["message_type"] == "" {
		return ""
	}
	message := time.Unix(0, event.Fields["timestamp"].(int64)*int64(time.Nanosecond)).String() + "\t" + event.Fields["message_type"].(string) + "\t" + event.Msg + "\n"
	buf.WriteString(message)

	return buf.String()
}

func (s *SumoLogicAppender) AppendLogs() {
	// the appender calls for the next message in the queue and parse it to a string
	s.logStringToSend.Write([]byte(StringBuilder(s.nozzleQueue.Pop())))
	s.logEventsInCurrentBuffer++
}

func (s *SumoLogicAppender) SendToSumo(log *bytes.Buffer) {
	logging.Trace.Println("Sending logs to Sumologic...")
	request, err := http.NewRequest("POST", s.url, log)
	if err != nil {
		logging.Error.Printf("http.NewRequest() error: %v\n", err)
		return
	}
	//request.Header.Add("content-type", "application/json")
	//request.SetBasicAuth("admin", "admin")
	response, err := s.httpClient.Do(request)

	if err != nil {
		logging.Error.Printf("http.Do() error: %v\n", err)
		return
	} else {
		logging.Trace.Println("Do(Request) successful")
	}
	s.logEventsInCurrentBuffer = 0                // reset counter
	s.logStringToSend = bytes.NewBufferString("") //reset String
	defer response.Body.Close()

}
