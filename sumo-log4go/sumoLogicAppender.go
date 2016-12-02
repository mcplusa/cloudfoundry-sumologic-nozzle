package sumoLogicAppender

import (
	"bufio"
	"bytes"
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

func (s *SumoLogicAppender) append(Event map[string]interface{}, Message string) {
	Event["msg"] = Message
	var buffer bytes.Buffer
	buffer.WriteString(event)
	reader := bufio.NewReader(resp.Body)
	for {
		line, err := reader.ReadBytes('\n')
		//...
		//log.Println(string(line))
	}

	sendToSumo(buffer.String())
}

func (boolean) checkEntryConditions() {
	if httpClient == nil {
		return false
	}
	return true
}

func senToSumo(log string) {
	if !checkEntryConditions() {
		return
	}
	request, err := http.NewRequest("POST", url, bytes.NewBufferString(log))
	//request.SetBasicAuth("admin", "admin")
	response, err := httpClient.Do(request)
	if !err {
		jghjk
		if response.StatusCode != 200 {
			//log print "Received HTTP error from Sumo Service:"
		}
		//consume the body if you want to re-use the connection
	} else {
		log.Fatal(err)
	}
	defer response.Body.Close()

}
