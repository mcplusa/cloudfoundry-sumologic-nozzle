package sumoLogicAppender

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"
)

type SumoLogicAppender struct {
	url               string
	connectionTimeout int //10000
	httpClient        http.Client
}

func newSumoLogicAppender(urlValue string, connectionTimeoutValue int) *SumoLogicAppender {
	return &SumoLogicAppender{
		url:               urlValue,
		connectionTimeout: connectionTimeoutValue,
		httpClient:        http.Client{Timeout: time.Duration(connectionTimeoutValue * int(time.Millisecond))},
	}
}

func (s *SumoLogicAppender) appendToSumo(Event map[string]interface{}, Message string) {
	//adding the message to the map
	Event["msg"] = Message
	jsonEvent, err := json.Marshal(Event)
	request, err := http.NewRequest("POST", s.url, bytes.NewBuffer(jsonEvent))
	response, err := s.httpClient.Do(request)
	if err != nil {
		if response.StatusCode != 200 {
			//log print "Received HTTP error from Sumo Service:"
		}
		//consume the body if you want to re-use the connection
	}
	defer response.Body.Close()
}
