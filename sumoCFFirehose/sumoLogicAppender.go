package sumoCFFirehose

import (
	"bytes"
	"compress/gzip"
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
	}
}

func (s *SumoLogicAppender) Start() {
	s.timerBetweenPost = time.Now()
	runtime.GOMAXPROCS(1)
	Buffer := newBuffer()
	Buffer.timerIdlebuffer = time.Now()
	logging.Info.Println("Starting Appender Worker")
	for {
		logging.Info.Println("Log queue size: ")
		logging.Info.Println(s.nozzleQueue.GetCount())
		if s.nozzleQueue.GetCount() == 0 {
			logging.Trace.Println("Waiting for 300 ms")
			time.Sleep(300 * time.Millisecond)
		}

		if time.Since(Buffer.timerIdlebuffer).Seconds() >= 10 && Buffer.logEventsInCurrentBuffer > 0 {
			logging.Info.Println("Sending current batch of logs after timer exceeded limit")
			go s.SendToSumo(&Buffer)
			Buffer = newBuffer()
			Buffer.timerIdlebuffer = time.Now()
			continue
		}

		if s.nozzleQueue.GetCount() != 0 {
			queueCount := s.nozzleQueue.GetCount()
			remainingBufferCount := s.eventsBatchSize - Buffer.logEventsInCurrentBuffer
			if queueCount >= remainingBufferCount {
				logging.Trace.Println("Pushing Logs to Sumo: ")
				logging.Trace.Println(remainingBufferCount)
				for i := 0; i < remainingBufferCount; i++ {
					s.AppendLogs(&Buffer)
					Buffer.timerIdlebuffer = time.Now()
				}
				go s.SendToSumo(&Buffer)
				Buffer = newBuffer()
			} else {
				logging.Trace.Println("Pushing Logs to Buffer: ")
				logging.Trace.Println(queueCount)
				for i := 0; i < queueCount; i++ {
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
	var buf bytes.Buffer
	g := gzip.NewWriter(&buf)
	g.Write(buffer.logStringToSend.Bytes())
	g.Close()

	for time.Since(s.timerBetweenPost) < s.sumoPostMinimumDelay {
		logging.Trace.Println("Delaying post to honor minimum post delay")
		time.Sleep(100 * time.Millisecond)
	}

	logging.Info.Println("Sending logs to Sumo Logic...")
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
		logging.Trace.Println("Post of logs successful")
		s.timerBetweenPost = time.Now()
	}

	defer response.Body.Close()
}
