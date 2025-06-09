package server

import (
	"encoding/json"
	"fmt"
	"mq/cmd/runtime"
)

func (mq *MQ) handleStrictjsRun(id string, data MQData) {

}

func (mq *MQ) handleStrictjsStop(id string, data MQData) {

}

func (mq *MQ) handleStrictjsUpdate(id string, data MQData) {

}

func (mq *MQ) handleStrictjsStatus(id string, data MQData) {

}

func (mq *MQ) handleScriptJsAdd(id string, data MQData) {
	fmt.Println(data)
	jsData := runtime.RuntimeData{}
	json.Unmarshal([]byte(data.Payload), &jsData)

	mq.Send(id, MQData{
		Cmd:       "S_ADD",
		ReplayId:  id,
		RequestId: data.RequestId,
		Topic:     data.Topic,
		Error:     "",
		Payload:   "",
	})

}

func (mq *MQ) handleScriptJsDel(id string, data MQData) {
	fmt.Println(data)
}
