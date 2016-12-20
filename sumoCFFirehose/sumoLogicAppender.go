package sumoCFFirehose

import (
	"bytes"
	"fmt"
	"net/http"
	"runtime"
	"time"

	"bitbucket.org/mcplusa-ondemand/firehose-to-sumologic/eventQueue"
	"bitbucket.org/mcplusa-ondemand/firehose-to-sumologic/events"
	"bitbucket.org/mcplusa-ondemand/firehose-to-sumologic/logging"
)

type SumoLogicAppender struct {
	url                  string
	connectionTimeout    int //10000
	httpClient           http.Client
	nozzleQueue          *eventQueue.Queue
	eventsBatchSize      int
	sumoPostMinimumDelay time.Duration
	timerBetweenPost     time.Time
}

type SumoBuffer struct {
	logStringToSend          *bytes.Buffer
	logEventsInCurrentBuffer int
	timerIdlebuffer          time.Time
	channelMessage           chan string
}

func NewSumoLogicAppender(urlValue string, connectionTimeoutValue int, nozzleQueue *eventQueue.Queue, eventsBatchSize int, sumoPostMinimumDelay time.Duration) *SumoLogicAppender {
	return &SumoLogicAppender{
		url:                  urlValue,
		connectionTimeout:    connectionTimeoutValue,
		httpClient:           http.Client{Timeout: time.Duration(connectionTimeoutValue * int(time.Millisecond))},
		nozzleQueue:          nozzleQueue,
		eventsBatchSize:      eventsBatchSize,
		sumoPostMinimumDelay: sumoPostMinimumDelay,
	}
}

func newBuffer() SumoBuffer {
	return SumoBuffer{
		logStringToSend:          bytes.NewBufferString(""),
		logEventsInCurrentBuffer: 0,
		channelMessage:           make(chan string),
	}
}

func (s *SumoLogicAppender) Start() {
	s.timerBetweenPost = time.Now()
	runtime.GOMAXPROCS(1)
	Buffer := newBuffer()
	fmt.Printf("First Buffer Reference: %v \n", &Buffer)
	Buffer.timerIdlebuffer = time.Now()
	msgFromChannel := ""
	logging.Info.Println("Starting Appender Worker")
	for {
		time.Sleep(300 * time.Millisecond)
		if Buffer.logEventsInCurrentBuffer >= s.eventsBatchSize || msgFromChannel == "Buffer being sent" {
			Buffer = newBuffer()
		}

		for s.nozzleQueue.GetCount() != 0 && Buffer.logEventsInCurrentBuffer < s.eventsBatchSize {
			s.AppendLogs(&Buffer)
			time.Sleep(300 * time.Millisecond)
			Buffer.timerIdlebuffer = time.Now()
			if Buffer.logEventsInCurrentBuffer == s.eventsBatchSize {
				logging.Info.Println("Batch Size complete")
				break
			} else if time.Since(Buffer.timerIdlebuffer).Seconds() >= 10 {
				logging.Info.Println("Sending current batch of logs after timer exceeded limit")
				break
			}
		}
		//wait period between posts to Sumo
		for time.Since(s.timerBetweenPost) < s.sumoPostMinimumDelay {
			time.Sleep(30 * time.Millisecond) // wait to retry
		}
		go s.SendToSumo(&Buffer)
		msgFromChannel = <-Buffer.channelMessage

		Buffer.timerIdlebuffer = time.Now() //reset Buffer timer
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

func (s *SumoLogicAppender) AppendLogs(buffer *SumoBuffer) {
	buffer.logStringToSend.Write([]byte(StringBuilder(s.nozzleQueue.Pop())))
	buffer.logEventsInCurrentBuffer++

}

func (s *SumoLogicAppender) SendToSumo(buffer *SumoBuffer) {
	buffer.channelMessage <- "Buffer being sent"

	logging.Trace.Println("Sending logs to Sumologic...")
	request, err := http.NewRequest("POST", s.url, buffer.logStringToSend)
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
		s.timerBetweenPost = time.Now() //reset timer post minimum
	}

	defer response.Body.Close()
}
