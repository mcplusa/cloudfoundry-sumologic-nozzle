package sumoCFFirehose

type SumoCFFirehose interface {
	Connect() bool
	AppendLogs(map[string]interface{}, string)
}
