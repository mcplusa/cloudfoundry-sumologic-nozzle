package sumoLog4go

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"time"
)

type SumoLogicAppender struct {
	url               string
	connectionTimeout int //10000
	httpClient        http.Client
}

func NewSumoLogicAppender(urlValue string, connectionTimeoutValue int) *SumoLogicAppender {
	return &SumoLogicAppender{
		url:               urlValue,
		connectionTimeout: connectionTimeoutValue,
		httpClient:        http.Client{Timeout: time.Duration(connectionTimeoutValue * int(time.Millisecond))},
	}
}

func (s *SumoLogicAppender) Connect() bool {
	success := false
	if s.url != "" {
		conn, err := net.Dial("tcp", s.url)
		fmt.Printf(fmt.Sprintf("Unable to connect to sumo server [%s]!\n", s.url), err.Error())
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

func (s *SumoLogicAppender) AppendLogs(Event map[string]interface{}, Message string) {
	//adding the message to the map
	Event["msg"] = Message
	jsonEvent, err := json.Marshal(Event)
	if err == nil {
		fmt.Println("-----here are the logs to send to sumo-------")

		//fmt.Println(string(jsonEvent)) //**
		s.SendToSumo(jsonEvent)
		//fmt.Println("---------------------------------------------")
	}

}

func (s *SumoLogicAppender) SendToSumo(log []byte) {
	request, err := http.NewRequest("POST", s.url, bytes.NewBuffer(log))
	if err != nil {
		fmt.Printf("http.NewRequest() error: %v\n", err)
		return
	}
	request.Header.Add("content-type", "application/json")
	//request.SetBasicAuth("admin", "admin")
	response, err := s.httpClient.Do(request)
	if err != nil {
		fmt.Printf("http.Do() error: %v\n", err)
		return
		//consume the body if you want to re-use the connection
	} else {
		fmt.Println("Do(Request) successful")
	}
	fmt.Println(response)
	defer response.Body.Close()

}
