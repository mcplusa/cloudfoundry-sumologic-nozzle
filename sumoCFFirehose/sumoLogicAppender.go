package sumoCFFirehose

import (
	"bytes"
	"fmt"
	"net/http"
	"runtime"
	"sync"
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
}

type SumoBuffer struct {
	logStringToSend          *bytes.Buffer
	logEventsInCurrentBuffer int
	timerPostMinimum         time.Time
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
		timerPostMinimum:         time.Now(),
	}
}

func (s *SumoLogicAppender) Start() {
	runtime.GOMAXPROCS(1)
	timer := time.Now()
	Buffer := newBuffer()
	var mutex = &sync.Mutex{} //synchronize access to Buffer
	logging.Info.Println("Starting Appender Worker")

	for {
		time.Sleep(300 * time.Millisecond)                        //delay
		if Buffer.logEventsInCurrentBuffer >= s.eventsBatchSize { //if buffer is full, create a new one
			fmt.Println("Creating new Buffer")
			Buffer = newBuffer()
		}
		mutex.Lock() //lock mutex to ensure exclusive access to buffer
		// while queue is not empty && s.eventsBatchSize not completed, queue.POP (appendLogs)
		for s.nozzleQueue.GetCount() != 0 && Buffer.logEventsInCurrentBuffer < s.eventsBatchSize {
			s.AppendLogs(&Buffer)                                     //this method POP an event from queue to Buffer
			timer = time.Now()                                        //reset timer
			if Buffer.logEventsInCurrentBuffer == s.eventsBatchSize { //if buffer is full, send logs to sumo
				logging.Info.Println("Batch Size complete")
				break
			} else if time.Since(timer).Seconds() >= 10 { // else if timer is up, send existing logs to sumo
				logging.Info.Println("Sending current batch of logs after timer exceeded limit")
				break
			}
		}
		//if batch size is met, send to sumo and reset temp buffer
		go s.SendToSumo(&Buffer)
		mutex.Unlock()
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

}

func (s *SumoLogicAppender) SendToSumo(buffer *SumoBuffer) {
	//wait period between posts to Sumo
	fmt.Println(buffer.logStringToSend.String())
	fmt.Println("......................")
	for time.Since(buffer.timerPostMinimum) < s.sumoPostMinimumDelay {
		time.Sleep(30 * time.Millisecond) // wait to retry
	}
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
		buffer.timerPostMinimum = time.Now() //reset timer post minimum
	}

	defer response.Body.Close()

}
