package sumoLog4go

type SumoLog4go interface {
	Connect() bool
	AppendLogs(map[string]interface{})
}
