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
	url                      string
	connectionTimeout        int //10000
	httpClient               http.Client
	nozzleQueue              eventQueue.Queue
	eventsBatchSize          int
	logEventsInCurrentBuffer int
	logStringToSend          string
}

func NewSumoLogicAppender(urlValue string, connectionTimeoutValue int, nozzleQueue eventQueue.Queue, eventsBatchSize int) *SumoLogicAppender {
	return &SumoLogicAppender{
		url:                      urlValue,
		connectionTimeout:        connectionTimeoutValue,
		httpClient:               http.Client{Timeout: time.Duration(connectionTimeoutValue * int(time.Millisecond))},
		nozzleQueue:              nozzleQueue,
		eventsBatchSize:          eventsBatchSize,
		logEventsInCurrentBuffer: 0,
		logStringToSend:          "",
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
	timer := time.NewTimer(60 * time.Second)
	fmt.Println("Starting Appender Worker")

	s.logStringToSend = ""
	for {
		fmt.Println("nozzle queue")
		fmt.Println(s.nozzleQueue.GetCount())
		time.Sleep(300 * time.Millisecond)
		if s.nozzleQueue.GetCount() != 0 {
			fmt.Println("i'm on the if")
			s.AppendLogs()
		}
		fmt.Println("passed first if")
		if s.logEventsInCurrentBuffer >= s.eventsBatchSize {
			s.SendToSumo(s.logStringToSend)
			s.logEventsInCurrentBuffer = 0 // reset counter
			s.logStringToSend = ""         //reset String
		} else if (<-timer.C).Second() == 0 {
			s.SendToSumo(s.logStringToSend)
			s.logEventsInCurrentBuffer = 0 // reset counter
			s.logStringToSend = ""         //reset String
		}
		fmt.Println("passed second if")
		fmt.Println("end of bucle")

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
	s.logStringToSend = s.logStringToSend + StringBuilder(s.nozzleQueue.Pop())
	s.logEventsInCurrentBuffer++

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
