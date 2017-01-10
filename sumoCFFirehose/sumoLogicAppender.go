package sumoCFFirehose

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
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
	sumoCategory         string
	sumoName             string
	sumoHost             string
	verboseLogMessages   bool
	customMetadata       string
}

type SumoBuffer struct {
	logStringToSend          *bytes.Buffer
	logEventsInCurrentBuffer int
	timerIdlebuffer          time.Time
}

func NewSumoLogicAppender(urlValue string, connectionTimeoutValue int, nozzleQueue *eventQueue.Queue, eventsBatchSize int, sumoPostMinimumDelay time.Duration, sumoCategory string, sumoName string, sumoHost string, verboseLogMessages bool, customMetadata string) *SumoLogicAppender {
	return &SumoLogicAppender{
		url:                  urlValue,
		connectionTimeout:    connectionTimeoutValue,
		httpClient:           http.Client{Timeout: time.Duration(connectionTimeoutValue * int(time.Millisecond))},
		nozzleQueue:          nozzleQueue,
		eventsBatchSize:      eventsBatchSize,
		sumoPostMinimumDelay: sumoPostMinimumDelay,
		sumoCategory:         sumoCategory,
		sumoName:             sumoName,
		sumoHost:             sumoHost,
		verboseLogMessages:   verboseLogMessages,
		customMetadata:       customMetadata,
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
	Buffer := newBuffer()
	Buffer.timerIdlebuffer = time.Now()
	logging.Info.Println("Starting Appender Worker")
	for {
		logging.Info.Printf("Log queue size: %d", s.nozzleQueue.GetCount())
		if s.nozzleQueue.GetCount() == 0 {
			logging.Trace.Println("Waiting for 300 ms")
			time.Sleep(300 * time.Millisecond)
		}

		if time.Since(Buffer.timerIdlebuffer).Seconds() >= 10 && Buffer.logEventsInCurrentBuffer > 0 {
			logging.Info.Println("Sending current batch of logs after timer exceeded limit")
			//wg.Add(1)
			go s.SendToSumo(Buffer.logStringToSend.String())
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

				go s.SendToSumo(Buffer.logStringToSend.String())
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

func StringBuilder(event *events.Event, verboseLogMessages bool) string {
	eventType := event.Type
	var msg []byte
	switch eventType {
	case "HttpStart":
		timestamp := time.Unix(0, event.Fields["timestamp"].(int64)*int64(time.Nanosecond)).String()
		event.Fields["timestamp"] = timestamp
		message, err := json.Marshal(event)
		if err == nil {
			msg = message
		}
	case "HttpStop":
		timestamp := time.Unix(0, event.Fields["timestamp"].(int64)*int64(time.Nanosecond)).String()
		event.Fields["timestamp"] = timestamp
		message, err := json.Marshal(event)
		if err == nil {
			msg = message
		}
	case "HttpStartStop":
		start_timestamp := time.Unix(0, event.Fields["start_timestamp"].(int64)*int64(time.Nanosecond)).String()
		event.Fields["start_timestamp"] = start_timestamp
		stop_timestamp := time.Unix(0, event.Fields["stop_timestamp"].(int64)*int64(time.Nanosecond)).String()
		event.Fields["stop_timestamp"] = stop_timestamp
		message, err := json.Marshal(event)
		if err == nil {
			msg = message
		}
	case "LogMessage":
		timestamp := time.Unix(0, event.Fields["timestamp"].(int64)*int64(time.Nanosecond)).String()
		event.Fields["timestamp"] = timestamp
		if verboseLogMessages == true {
			message, err := json.Marshal(event)
			if err == nil {
				msg = message
			}
		} else {
			eventNoVerbose := events.Event{
				Fields: map[string]interface{}{
					"timestamp":   event.Fields["timestamp"],
					"cf_app_guid": event.Fields["cf_app_id"],
				},
				Msg:  event.Msg,
				Type: event.Type,
			}
			message, err := json.Marshal(eventNoVerbose)
			if err == nil {
				msg = message
			}
		}
	case "ValueMetric":
		message, err := json.Marshal(event)
		if err == nil {
			msg = message
		}
	case "CounterEvent":
		message, err := json.Marshal(event)
		if err == nil {
			msg = message
		}
	case "Error":
		message, err := json.Marshal(event)
		if err == nil {
			msg = message
		}
	case "ContainerMetric":
		message, err := json.Marshal(event)
		if err == nil {
			msg = message
		}
	}
	buf := new(bytes.Buffer)
	buf.Write(msg)
	return buf.String() + "\n"
}

func (s *SumoLogicAppender) AppendLogs(buffer *SumoBuffer) {
	buffer.logStringToSend.Write([]byte(StringBuilder(s.nozzleQueue.Pop(), s.verboseLogMessages)))
	buffer.logEventsInCurrentBuffer++

}
func ParseCustomMetadata(customMetadata string) map[string]string {
	cMetadataArray := strings.Split(customMetadata, ",")
	customMetadataMap := make(map[string]string)
	for i := 0; i < len(cMetadataArray); i++ {
		customMetadataMap[strings.Split(cMetadataArray[i], ":")[0]] = strings.Split(cMetadataArray[i], ":")[1]
	}
	return customMetadataMap
}

func (s *SumoLogicAppender) SendToSumo(logStringToSend string) {
	if logStringToSend != "" {
		var buf bytes.Buffer
		g := gzip.NewWriter(&buf)
		g.Write([]byte(logStringToSend))
		g.Close()
		request, err := http.NewRequest("POST", s.url, &buf)
		if err != nil {
			logging.Error.Printf("http.NewRequest() error: %v\n", err)
			return
		}
		request.Header.Add("Content-Encoding", "gzip")

		if s.sumoName != "" {
			request.Header.Add("X-Sumo-Name", s.sumoName)
		}
		if s.sumoHost != "" {
			request.Header.Add("X-Sumo-Host", s.sumoHost)
		}
		if s.sumoCategory != "" {
			request.Header.Add("X-Sumo-Category", s.sumoCategory)
		}

		if s.customMetadata != "" {
			customMetadataMap := ParseCustomMetadata(s.customMetadata)
			for key, value := range customMetadataMap {
				request.Header.Add(key, value)
			}
		}
		//checking the timer before first POST intent
		for time.Since(s.timerBetweenPost) < s.sumoPostMinimumDelay {
			logging.Trace.Println("Delaying Post because minimum post timer not expired")
			time.Sleep(100 * time.Millisecond)
		}
		response, err := s.httpClient.Do(request)

		if (err != nil) || (response.StatusCode != 200 && response.StatusCode != 302 && response.StatusCode < 500) {
			logging.Info.Println("Endpoint dropped the post send")
			logging.Info.Println("Waiting for 300 ms to retry")
			time.Sleep(300 * time.Millisecond)
			statusCode := 0
			err := Retry(func(attempt int) (bool, error) {
				var errRetry error
				request, err := http.NewRequest("POST", s.url, &buf)
				if err != nil {
					logging.Error.Printf("http.NewRequest() error: %v\n", err)
				}
				request.Header.Add("Content-Encoding", "gzip")

				if s.sumoName != "" {
					request.Header.Add("X-Sumo-Name", s.sumoName)
				}
				if s.sumoHost != "" {
					request.Header.Add("X-Sumo-Host", s.sumoHost)
				}
				if s.sumoCategory != "" {
					request.Header.Add("X-Sumo-Category", s.sumoCategory)
				}
				//checking the timer before POST (retry intent)
				for time.Since(s.timerBetweenPost) < s.sumoPostMinimumDelay {
					logging.Trace.Println("Delaying Post because minimum post timer not expired")
					time.Sleep(100 * time.Millisecond)
				}
				response, errRetry = s.httpClient.Do(request)
				if errRetry != nil {
					logging.Error.Printf("http.Do() error: %v\n", errRetry)
					logging.Info.Println("Waiting for 300 ms to retry after error")
					time.Sleep(300 * time.Millisecond)
					return attempt < 5, errRetry
				} else if response.StatusCode != 200 && response.StatusCode != 302 && response.StatusCode < 500 {
					logging.Info.Println("Endpoint dropped the post send again")
					logging.Info.Println("Waiting for 300 ms to retry after a retry ...")
					statusCode = response.StatusCode
					time.Sleep(300 * time.Millisecond)
					return attempt < 5, errRetry
				} else if response.StatusCode == 200 {
					logging.Trace.Println("Post of logs successful after retry...")
					s.timerBetweenPost = time.Now()
					statusCode = response.StatusCode
					return true, err
				}
				return attempt < 5, errRetry
			})
			if err != nil {
				logging.Error.Println("Error, Not able to post after retry")
				logging.Error.Printf("http.Do() error: %v\n", err)
				return
			} else if statusCode != 200 {
				logging.Error.Printf("Not able to post after retry, with status code: %d", statusCode)
			}
		} else if response.StatusCode == 200 {
			logging.Trace.Println("Post of logs successful")
			s.timerBetweenPost = time.Now()
		}

		if response != nil {
			defer response.Body.Close()
		}
	}

}

//------------------Retry Logic Code-------------------------------

// MaxRetries is the maximum number of retries before bailing.
var MaxRetries = 10
var errMaxRetriesReached = errors.New("exceeded retry limit")

// Func represents functions that can be retried.
type Func func(attempt int) (retry bool, err error)

// Do keeps trying the function until the second argument
// returns false, or no error is returned.
func Retry(fn Func) error {
	var err error
	var cont bool
	attempt := 1
	for {
		cont, err = fn(attempt)
		if !cont || err == nil {
			break
		}
		attempt++
		if attempt > MaxRetries {
			return errMaxRetriesReached
		}
	}
	return err
}

// IsMaxRetries checks whether the error is due to hitting the
// maximum number of retries or not.
func IsMaxRetries(err error) bool {
	return err == errMaxRetriesReached
}
