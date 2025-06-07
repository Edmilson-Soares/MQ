package mq

import (
	"encoding/json"
	"strings"
)

func (mq *MQ) handleGet(id string, data MQData) {
	key := data.Topic
	bucket := "store"
	if strings.Contains(data.Topic, ":") {
		bucket = strings.Split(data.Topic, ":")[0]
		key = strings.Split(data.Topic, ":")[1]
	}
	str, err := mq.KV.BGet(bucket, key)
	if err != nil {
		mq.Send(id, MQData{
			Cmd:       "GET",
			ReplayId:  id,
			RequestId: data.RequestId,
			Topic:     data.Topic,
			Error:     err.Error(),
		})
		return
	}
	mq.Send(id, MQData{
		Cmd:       "GET",
		ReplayId:  id,
		RequestId: data.RequestId,
		Topic:     data.Topic,
		Payload:   str,
	})

}

func (mq *MQ) handleSet(id string, data MQData) {
	key := data.Topic
	bucket := "store"
	if strings.Contains(data.Topic, ":") {
		bucket = strings.Split(data.Topic, ":")[0]
		key = strings.Split(data.Topic, ":")[1]
	}
	err := mq.KV.BSet(bucket, key, data.Payload)
	if err != nil {
		mq.Send(id, MQData{
			Cmd:       "SET",
			ReplayId:  id,
			RequestId: data.RequestId,
			Topic:     data.Topic,
			Error:     err.Error(),
		})
		return
	}

	mq.Send(id, MQData{
		Cmd:       "SET",
		ReplayId:  id,
		RequestId: data.RequestId,
		Topic:     data.Topic,
		Payload:   "ok",
	})

}

func (mq *MQ) handleDel(id string, data MQData) {
	key := data.Topic
	bucket := "store"
	if strings.Contains(data.Topic, ":") {
		bucket = strings.Split(data.Topic, ":")[0]
		key = strings.Split(data.Topic, ":")[1]
	}
	err := mq.KV.BDel(bucket, key)
	if err != nil {
		mq.Send(id, MQData{
			Cmd:       "DEL",
			ReplayId:  id,
			RequestId: data.RequestId,
			Topic:     data.Topic,
			Error:     err.Error(),
		})
		return
	}
	mq.Send(id, MQData{
		Cmd:       "DEL",
		ReplayId:  id,
		RequestId: data.RequestId,
		Topic:     data.Topic,
		Payload:   "ok",
	})
}

func (mq *MQ) handleBDel(id string, data MQData) {

	err := mq.KV.BDelete(data.Topic)
	if err != nil {
		mq.Send(id, MQData{
			Cmd:       "BDEL",
			ReplayId:  id,
			RequestId: data.RequestId,
			Topic:     data.Topic,
			Error:     err.Error(),
		})
		return
	}
	mq.Send(id, MQData{
		Cmd:       "BDEL",
		ReplayId:  id,
		RequestId: data.RequestId,
		Topic:     data.Topic,
		Payload:   "",
	})
}

func (mq *MQ) handleBAdd(id string, data MQData) {

	err := mq.KV.BCreate(data.Topic)
	if err != nil {
		mq.Send(id, MQData{
			Cmd:       "BADD",
			ReplayId:  id,
			RequestId: data.RequestId,
			Topic:     data.Topic,
			Error:     err.Error(),
		})
		return
	}
	mq.Send(id, MQData{
		Cmd:       "BADD",
		ReplayId:  id,
		RequestId: data.RequestId,
		Topic:     data.Topic,
		Payload:   "ok",
	})
}

//////////////////////////

func (mq *MQ) handleBFilterKey(id string, data MQData) {

	bucket := data.Topic
	if bucket == "" {
		bucket = "store"
	}
	res, err := mq.KV.BList(bucket, func(k, v []byte) bool {
		return strings.HasPrefix(string(k), data.Payload)
	})

	if err != nil {
		mq.Send(id, MQData{
			Cmd:       "BFK",
			ReplayId:  id,
			RequestId: data.RequestId,
			Topic:     data.Topic,
			Error:     err.Error(),
		})
		return
	}

	str, _ := json.Marshal(res)
	mq.Send(id, MQData{
		Cmd:       "BFK",
		ReplayId:  id,
		RequestId: data.RequestId,
		Topic:     data.Topic,
		Payload:   string(str),
	})
}

func (mq *MQ) handleBFilterVal(id string, data MQData) {

	bucket := data.Topic
	if bucket == "" {
		bucket = "store"
	}
	res, err := mq.KV.BList(bucket, func(k, v []byte) bool {
		return strings.HasPrefix(string(v), data.Payload)
	})

	if err != nil {
		mq.Send(id, MQData{
			Cmd:       "BFV",
			ReplayId:  id,
			RequestId: data.RequestId,
			Topic:     data.Topic,
			Error:     err.Error(),
			Payload:   "",
		})
		return
	}

	str, _ := json.Marshal(res)
	mq.Send(id, MQData{
		Cmd:       "BFV",
		ReplayId:  id,
		RequestId: data.RequestId,
		Topic:     data.Topic,
		Error:     "",
		Payload:   string(str),
	})
}
