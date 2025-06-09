package server

import (
	"encoding/json"
	"mq/cmd/db"
)

func (mq *MQ) handledbCreateCollection(id string, data MQData) {
	err := mq.DB.CreateIndex(data.Topic, data.Payload)
	if err != nil {
		mq.Send(id, MQData{
			Cmd:       "DB_CC",
			ReplayId:  id,
			RequestId: data.RequestId,
			Topic:     data.Topic,
			Error:     err.Error(),
			Payload:   "",
		})
		return
	}

	mq.Send(id, MQData{
		Cmd:       "DB_CC",
		ReplayId:  id,
		RequestId: data.RequestId,
		Topic:     data.Topic,
		Error:     "",
		Payload:   "",
	})

}

// 1749476853203693700
func (mq *MQ) handledbDeleteCollection(id string, data MQData) {
	err := mq.DB.RemoveIndex(data.Topic)
	if err != nil {
		mq.Send(id, MQData{
			Cmd:       "DB_CD",
			ReplayId:  id,
			RequestId: data.RequestId,
			Topic:     data.Topic,
			Error:     err.Error(),
			Payload:   "",
		})
		return
	}

	mq.Send(id, MQData{
		Cmd:       "DB_CD",
		ReplayId:  id,
		RequestId: data.RequestId,
		Topic:     data.Topic,
		Error:     "",
		Payload:   "",
	})
}
func (mq *MQ) handledbInsertCollection(id string, data MQData) {
	doc := db.Document{}
	json.Unmarshal([]byte(data.Payload), &doc)
	id_, err := mq.DB.Insert(data.Topic, doc)
	if err != nil {
		mq.Send(id, MQData{
			Cmd:       "DB_CI",
			ReplayId:  id,
			RequestId: data.RequestId,
			Topic:     data.Topic,
			Error:     err.Error(),
			Payload:   "",
		})
		return
	}

	mq.Send(id, MQData{
		Cmd:       "DB_CI",
		ReplayId:  id,
		RequestId: data.RequestId,
		Topic:     data.Topic,
		Error:     "",
		Payload:   id_,
	})
}
func (mq *MQ) handledbRemoveCollection(id string, data MQData) {
	err := mq.DB.Delete(data.Topic, data.Payload)
	if err != nil {
		mq.Send(id, MQData{
			Cmd:       "DB_CR",
			ReplayId:  id,
			RequestId: data.RequestId,
			Topic:     data.Topic,
			Error:     err.Error(),
			Payload:   "",
		})
		return
	}
	mq.Send(id, MQData{
		Cmd:       "DB_CR",
		ReplayId:  id,
		RequestId: data.RequestId,
		Topic:     data.Topic,
		Error:     "",
		Payload:   data.RequestId,
	})
}
func (mq *MQ) handledbUpdateCollection(id string, data MQData) {
	doc := db.Document{}
	json.Unmarshal([]byte(data.Payload), &doc)
	id_ := doc["_id"].(string)
	err := mq.DB.Update(data.Topic, id_, doc)
	if err != nil {
		mq.Send(id, MQData{
			Cmd:       "DB_CU",
			ReplayId:  id,
			RequestId: data.RequestId,
			Topic:     data.Topic,
			Error:     err.Error(),
			Payload:   "",
		})
		return
	}

	mq.Send(id, MQData{
		Cmd:       "DB_CU",
		ReplayId:  id,
		RequestId: data.RequestId,
		Topic:     data.Topic,
		Error:     "",
		Payload:   id_,
	})
}
func (mq *MQ) handledbFilterCollection(id string, data MQData) {
	query := map[string]interface{}{}
	json.Unmarshal([]byte(data.Payload), &query)
	results, err := mq.DB.FindWithQuery(data.Topic, query)
	if err != nil {
		mq.Send(id, MQData{
			Cmd:       "DB_CF",
			ReplayId:  id,
			RequestId: data.RequestId,
			Topic:     data.Topic,
			Error:     err.Error(),
			Payload:   "",
		})
		return
	}
	strInput, _ := json.Marshal(results)
	mq.Send(id, MQData{
		Cmd:       "DB_CF",
		ReplayId:  id,
		RequestId: data.RequestId,
		Topic:     data.Topic,
		Error:     "",
		Payload:   string(strInput),
	})

}
func (mq *MQ) handledbListCollection(id string, data MQData) {
	query := map[string]interface{}{}
	json.Unmarshal([]byte(data.Payload), &query)
	results, err := mq.DB.FindAll(data.Topic)
	if err != nil {
		mq.Send(id, MQData{
			Cmd:       "DB_CL",
			ReplayId:  id,
			RequestId: data.RequestId,
			Topic:     data.Topic,
			Error:     err.Error(),
			Payload:   "",
		})
		return
	}
	strInput, _ := json.Marshal(results)
	mq.Send(id, MQData{
		Cmd:       "DB_CL",
		ReplayId:  id,
		RequestId: data.RequestId,
		Topic:     data.Topic,
		Error:     "",
		Payload:   string(strInput),
	})

}
func (mq *MQ) handledbFindOneCollection(id string, data MQData) {

	doc, err := mq.DB.FindOne(data.Topic, data.Payload)
	if err != nil {
		mq.Send(id, MQData{
			Cmd:       "DB_CG",
			ReplayId:  id,
			RequestId: data.RequestId,
			Topic:     data.Topic,
			Error:     err.Error(),
			Payload:   "",
		})
		return
	}
	strInput, _ := json.Marshal(doc)
	mq.Send(id, MQData{
		Cmd:       "DB_CG",
		ReplayId:  id,
		RequestId: data.RequestId,
		Topic:     data.Topic,
		Error:     "",
		Payload:   string(strInput),
	})
}
