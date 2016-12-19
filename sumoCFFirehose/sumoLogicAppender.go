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
	url                      string
	connectionTimeout        int //10000
	httpClient               http.Client
	nozzleQueue              *eventQueue.Queue
	eventsBatchSize          int
	logEventsInCurrentBuffer int
	logStringToSend          *bytes.Buffer
	sumoPostMinimumDelay     time.Duration
	timerPostMinimum         time.Time
	bufferToSend             SumoBuffer
}

type SumoBuffer struct {
	logStringToSend          *bytes.Buffer
	logEventsInCurrentBuffer int
}

func NewSumoLogicAppender(urlValue string, connectionTimeoutValue int, nozzleQueue *eventQueue.Queue, eventsBatchSize int, sumoPostMinimumDelay time.Duration) *SumoLogicAppender {
	return &SumoLogicAppender{
		url:                      urlValue,
		connectionTimeout:        connectionTimeoutValue,
		httpClient:               http.Client{Timeout: time.Duration(connectionTimeoutValue * int(time.Millisecond))},
		nozzleQueue:              nozzleQueue,
		eventsBatchSize:          eventsBatchSize,
		logEventsInCurrentBuffer: 0,
		logStringToSend:          bytes.NewBufferString(""),
		sumoPostMinimumDelay:     sumoPostMinimumDelay,
		timerPostMinimum:         time.Now(),
		bufferToSend: SumoBuffer{
			logStringToSend:          bytes.NewBufferString(""),
			logEventsInCurrentBuffer: 0,
		},
	}
}

func (s *SumoLogicAppender) Start() {
	runtime.GOMAXPROCS(1)
	timer := time.Now()
	tempBuffer := &SumoBuffer{
		logStringToSend:          bytes.NewBufferString(""),
		logEventsInCurrentBuffer: 0,
	}

	logging.Info.Println("Starting Appender Worker")
	for {
		time.Sleep(300 * time.Millisecond) //delay
		// while queue is not empty && s.eventsBatchSize not completed, queue.POP (appendLogs)
		for s.nozzleQueue.GetCount() != 0 && s.bufferToSend.logEventsInCurrentBuffer < s.eventsBatchSize {
			s.AppendLogs(&s.bufferToSend)                                                                                                //this method POP an event from queue to Buffer
			timer = time.Now()                                                                                                           //reset timer
			if s.bufferToSend.logEventsInCurrentBuffer == s.eventsBatchSize && tempBuffer.logEventsInCurrentBuffer < s.eventsBatchSize { //if buffer is full, send logs to sumo
				logging.Info.Println("Batch Size complete, filling temp buffer")
				s.AppendLogs(tempBuffer)
				break
			} else if time.Since(timer).Seconds() >= 10 { // else if timer is up, send existing logs to sumo
				logging.Info.Println("Sending current batch of logs after timer exceeded limit")
				go s.SendToSumo(s.bufferToSend.logStringToSend)
				break
			}
		}
		//if batch size is met, send to sumo and reset temp buffer
		if s.bufferToSend.logEventsInCurrentBuffer == s.eventsBatchSize {
			go s.SendToSumo(s.bufferToSend.logStringToSend)
			s.bufferToSend.logStringToSend = tempBuffer.logStringToSend
			s.bufferToSend.logEventsInCurrentBuffer = tempBuffer.logEventsInCurrentBuffer
			tempBuffer = &SumoBuffer{ //reset temp buffer
				logStringToSend:          bytes.NewBufferString(""),
				logEventsInCurrentBuffer: 0,
			}
		}

		timer = time.Now() //reset timer
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
	// the appender calls for the next message in the queue and parse it to a string
	// then fills a buffer with the message
	buffer.logStringToSend.Write([]byte(StringBuilder(s.nozzleQueue.Pop())))
	buffer.logEventsInCurrentBuffer++
	fmt.Println(s.bufferToSend.logEventsInCurrentBuffer)

}

func (s *SumoLogicAppender) SendToSumo(log *bytes.Buffer) {
	//wait period between posts to Sumo
	fmt.Println(log.String())
	fmt.Println(".............")
	for time.Since(s.timerPostMinimum) < s.sumoPostMinimumDelay {
		time.Sleep(30 * time.Millisecond) // wait to retry
	}
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
		s.timerPostMinimum = time.Now() //reset timer post minimum
	}
	s.bufferToSend = SumoBuffer{
		logStringToSend:          bytes.NewBufferString(""), //reset String
		logEventsInCurrentBuffer: 0,                         //reset counter
	}
	defer response.Body.Close()

}
