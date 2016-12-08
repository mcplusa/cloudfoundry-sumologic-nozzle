package sumoCFFirehose

import (
	"bytes"
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

func (s *SumoLogicAppender) AppendLogs(fields map[string]interface{}, message string) {
	//adding the message to the map
	//fmt.Printf("I'm in AppendLogs method")
	/*if Event == nil {
		return
	}

	if Event["msg"] == nil {
		return
	}

	if Event["msg"] == "" {
		return
	}*/
	//timestampToInt, err := strconv.ParseInt(fields["timestamp"].(string), 10, 64)
	//if err == nil {
	//date := time.Unix(timestampToInt, 0)
	Message := /*strconv.Itoa(fields["timestamp"]) + */ "\t" + fields["message_type"].(string) + "\t" + message
	/*fmt.Println("---------------------I'm about to do an HTTP Post with this message-------------------")
	fmt.Println(Message)
	fmt.Println("---------------------------------------------------------------------------------------")*/
	s.SendToSumo(Message)
	//}

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
