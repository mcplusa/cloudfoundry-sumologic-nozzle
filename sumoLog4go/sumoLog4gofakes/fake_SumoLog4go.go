// This file was generated by counterfeiter
package sumoLog4gofakes

import (
	"sync"

	"bitbucket.org/mcplusa-ondemand/firehouse-to-sumologic/sumolog4go"
)

type FakeSumoLog4go struct {
	ConnectStub        func() bool
	connectMutex       sync.RWMutex
	connectArgsForCall []struct{}
	connectReturns     struct {
		result1 bool
	}
	AppendLogsStub        func(map[string]interface{}, string)
	AppendLogsMutex       sync.RWMutex
	AppendLogsArgsForCall []struct {
		arg1 map[string]interface{}
		arg2 string
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeSumoLog4go) Connect() bool {
	fake.connectMutex.Lock()
	fake.connectArgsForCall = append(fake.connectArgsForCall, struct{}{})
	fake.recordInvocation("Connect", []interface{}{})
	fake.connectMutex.Unlock()
	if fake.ConnectStub != nil {
		return fake.ConnectStub()
	} else {
		return fake.connectReturns.result1
	}
}

func (fake *FakeSumoLog4go) ConnectCallCount() int {
	fake.connectMutex.RLock()
	defer fake.connectMutex.RUnlock()
	return len(fake.connectArgsForCall)
}

func (fake *FakeSumoLog4go) ConnectReturns(result1 bool) {
	fake.ConnectStub = nil
	fake.connectReturns = struct {
		result1 bool
	}{result1}
}

func (fake *FakeSumoLog4go) AppendLogs(arg1 map[string]interface{}, arg2 string) {
	fake.AppendLogsMutex.Lock()
	fake.AppendLogsArgsForCall = append(fake.AppendLogsArgsForCall, struct {
		arg1 map[string]interface{}
		arg2 string
	}{arg1, arg2})
	fake.recordInvocation("ShipEvents", []interface{}{arg1, arg2})
	fake.AppendLogsMutex.Unlock()
	if fake.AppendLogsStub != nil {
		fake.AppendLogsStub(arg1, arg2)
	}
}

func (fake *FakeSumoLog4go) ShipEventsCallCount() int {
	fake.AppendLogsMutex.RLock()
	defer fake.AppendLogsMutex.RUnlock()
	return len(fake.AppendLogsArgsForCall)
}

func (fake *FakeSumoLog4go) ShipEventsArgsForCall(i int) (map[string]interface{}, string) {
	fake.AppendLogsMutex.RLock()
	defer fake.AppendLogsMutex.RUnlock()
	return fake.AppendLogsArgsForCall[i].arg1, fake.AppendLogsArgsForCall[i].arg2
}

func (fake *FakeSumoLog4go) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.connectMutex.RLock()
	defer fake.connectMutex.RUnlock()
	fake.AppendLogsMutex.RLock()
	defer fake.AppendLogsMutex.RUnlock()
	return fake.invocations
}

func (fake *FakeSumoLog4go) recordInvocation(key string, args []interface{}) {
	fake.invocationsMutex.Lock()
	defer fake.invocationsMutex.Unlock()
	if fake.invocations == nil {
		fake.invocations = map[string][][]interface{}{}
	}
	if fake.invocations[key] == nil {
		fake.invocations[key] = [][]interface{}{}
	}
	fake.invocations[key] = append(fake.invocations[key], args)
}

var _ sumoLog4go.SumoLog4go = new(FakeSumoLog4go)
