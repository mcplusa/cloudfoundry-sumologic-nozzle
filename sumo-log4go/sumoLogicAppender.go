package sumoLogicAppender

import(
	"net/http"
	"net/url"
    "bytes"
    "log"
    "io/ioutil"
    "time"
    "bitbucket.org/mcplusa-ondemand/firehouse-to-sumologic/eventRouting"

)

type sumoLogicAppender struct{
    string url = nil
    int connectionTimeout = 1000
    //int socketTimeout = 60000
    http.Client httpClient = nil
}

// setUrl receives a pointer to sumoLogicAppender so it can modify it.
func (s *sumoLogicAppender) setUrl(string url) {
    s.url = url
}

// connectionTimeout receives a pointer to sumoLogicAppender so it can modify it.
func (s *sumoLogicAppender) connectionTimeout(int connectionTimeout) {
    s.connectionTimeout = connectionTimeout
}

// socketTimeout receives a pointer to sumoLogicAppender so it can modify it.
func (s *sumoLogicAppender) socketTimeout(int socketTimeout) {
    s.socketTimeout = socketTimeout
}

func (boolean) isInitialized() {
    return httpClient != nil;
}

func activateOptions(){
    httpClient := {
        Timeout: connectionTimeout * time.Millisecod,
    }
}


func append(Event event){
    if !checkEntryConditions() {
        return
    }
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
    if httpClient==nil {
        return false
    }
    return true
}


func senToSumo(string log){
    if !checkEntryConditions() {
        return
    }
    request, err := http.NewRequest("POST", url, bytes.NewBufferString(log))
    //request.SetBasicAuth("admin", "admin")
    response, err := httpClient.Do(request)
    if !err {
       if response.StatusCode !=200 {
           //log print "Received HTTP error from Sumo Service:"
       }
       //consume the body if you want to re-use the connection
    }else{
        log.Fatal(err)
    }
    defer response.Body.Close()

}



