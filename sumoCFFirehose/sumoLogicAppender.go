package sumoCFFirehose

import (
	"bytes"
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
	Buffer.timerIdlebuffer = time.Now()
	logging.Info.Println("Starting Appender Worker")
	for {
		for s.nozzleQueue.GetCount() == 0 {
			time.Sleep(300 * time.Millisecond)
		}

		if time.Since(Buffer.timerIdlebuffer).Seconds() >= 10 && Buffer.logEventsInCurrentBuffer > 0 {
			logging.Info.Println("Sending current batch of logs after timer exceeded limit")
			go s.SendToSumo(&Buffer)
			<-Buffer.channelMessage
			Buffer = newBuffer()
		}
		for s.nozzleQueue.GetCount() != 0 {
			if s.nozzleQueue.GetCount() >= s.eventsBatchSize-Buffer.logEventsInCurrentBuffer {
				for Buffer.logEventsInCurrentBuffer < s.eventsBatchSize {
					s.AppendLogs(&Buffer)
					Buffer.timerIdlebuffer = time.Now()
				}
				logging.Trace.Println("Batch Size complete")
				go s.SendToSumo(&Buffer)
				<-Buffer.channelMessage
				Buffer = newBuffer()
			} else {
				for s.nozzleQueue.GetCount() > 0 && Buffer.logEventsInCurrentBuffer < s.eventsBatchSize {
					//TODO fill the buffer with whatever is in the queue without sending to sumo
					s.AppendLogs(&Buffer)
					Buffer.timerIdlebuffer = time.Now()
				}
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

func (s *SumoLogicAppender) AppendLogs(buffer *SumoBuffer) {
	buffer.logStringToSend.Write([]byte(StringBuilder(s.nozzleQueue.Pop())))
	buffer.logEventsInCurrentBuffer++

}

func (s *SumoLogicAppender) SendToSumo(buffer *SumoBuffer) {
	//wait period between posts to Sumo
	for time.Since(s.timerBetweenPost) < s.sumoPostMinimumDelay {
		time.Sleep(100 * time.Millisecond) // wait to retry
	}
	/*fmt.Println(buffer.logStringToSend.String())
	fmt.Println("........")*/
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
