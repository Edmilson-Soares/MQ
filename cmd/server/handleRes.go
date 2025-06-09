package server

func (mq *MQ) handleRes(id string, data MQData) {

	if data.ReplayId == "self" {
		mq.chrequest[data.RequestId] <- MQResponse{
			Payload: data.Payload,
			Error:   data.Error,
		}
		return
	}
	if mq.clients[data.ReplayId] != nil {
		mq.Send(data.ReplayId, MQData{
			Cmd:       "RES",
			ReplayId:  id,
			Error:     data.Error,
			RequestId: data.RequestId,
			Topic:     data.Topic,
			Payload:   data.Payload,
		})

	}
}
