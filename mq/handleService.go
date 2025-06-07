package mq

func (mq *MQ) handleService(id string, data MQData) {
	mq.services[data.Topic] = id
	mq.Send(id, MQData{
		Cmd:     "OK",
		Topic:   data.Topic,
		Payload: "",
	})
}
