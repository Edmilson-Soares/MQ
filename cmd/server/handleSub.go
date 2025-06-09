package server

func (mq *MQ) handleSub(id string, data MQData) {
	mq.subs[data.Topic] = append(mq.subs[data.Topic], id)
	mq.Send(id, MQData{
		Cmd:     "OK",
		Topic:   data.Topic,
		Payload: "",
	})
}
