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
	eventsBatchSize   int
}

func NewSumoLogicAppender(urlValue string, connectionTimeoutValue int, nozzleQueue eventQueue.Queue, eventsBatchSize int) *SumoLogicAppender {
	return &SumoLogicAppender{
		url:               urlValue,
		connectionTimeout: connectionTimeoutValue,
		httpClient:        http.Client{Timeout: time.Duration(connectionTimeoutValue * int(time.Millisecond))},
		nozzleQueue:       nozzleQueue,
		eventsBatchSize:   eventsBatchSize,
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

func (s *SumoLogicAppender) Start() {
	s.AppendLogs()
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
	logMessage := ""
	stringBuilderCalls := 0
	for s.nozzleQueue.GetCount() > s.eventsBatchSize { //when the batch limit is met, call stringBuilder
		logMessage = logMessage + StringBuilder(s.nozzleQueue.Pop())
		stringBuilderCalls++
		if s.eventsBatchSize == stringBuilderCalls {
			s.SendToSumo(logMessage)
			stringBuilderCalls = 0
			logMessage = ""
		}
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
