package mq

func (mq *MQ) handleReq(id string, data MQData) {

	if mq.services[data.Topic] != "" {
		req := mq.services[data.Topic]
		if req == "self" {
			go mq.serviceself[data.Topic](data, func(err string, payload string) {
				mq.handleRes(id, MQData{
					Cmd:       "RES",
					Topic:     data.Topic,
					Payload:   payload,
					FromId:    id,
					Error:     err,
					RequestId: data.RequestId,
					ReplayId:  id,
				})
			})

		} else {
			mq.Send(mq.services[data.Topic], MQData{
				Cmd:       "REQ",
				ReplayId:  id,
				RequestId: data.RequestId,
				Topic:     data.Topic,
				Payload:   data.Payload,
			})
		}

	}
}
