package server

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

func (mq *MQ) Publish(topic, payload string) {
	mq.handlePub(MQData{
		Topic:   topic,
		Payload: payload,
	})
}

func (mq *MQ) Subscribe(topic string, cb func(data MQData)) {
	mq.subs[topic] = append(mq.subs[topic], "self")
	mq.subself[topic] = append(mq.subself[topic], cb)
}
func (mq *MQ) Service(topic string, fn func(data MQData, replay func(err string, payload string))) {
	mq.services[topic] = "self"
	mq.serviceself[topic] = fn
}

func (mq *MQ) Request(topic, Payload string, timeout time.Duration) (string, error) {
	reqId := uuid.New().String()
	mq.chrequest[reqId] = make(chan MQResponse) // cria um canal de string
	mq.handleReq("self", MQData{
		Cmd:       "REQ",
		FromId:    "self",
		Topic:     topic,
		RequestId: reqId,
		Payload:   Payload,
	})
	ch, existe := mq.chrequest[reqId]
	if !existe {
		return "", fmt.Errorf("canal %s não existe", topic)
	}
	select {
	case res := <-ch:
		close(ch)
		if res.Error != "" {
			return "", errors.New("Error :" + res.Error)
		}
		return res.Payload, nil
	case <-time.After(timeout):
		close(ch)
		return "", fmt.Errorf("timeout de %v expirado no canal %s", timeout, topic)
	}
}
