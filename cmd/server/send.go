package server

import (
	"net"
	"strings"
)

func (mq *MQ) send(conn net.Conn, data MQData) error {
	str, err := structToJSON(data)
	if err != nil {
		return err
	}
	conn.Write([]byte(str + "\n"))
	return err
}

func (mq *MQ) Send(id string, data MQData) error {
	if id == "self" {
		topic := data.Topic
		if strings.Contains(data.Regtopic, ".*.") {
			topic = data.Regtopic
		}
		for _, fn := range mq.subself[topic] {
			go fn(data)
		}
		return nil
	}
	str, err := structToJSON(data)
	if err != nil {
		return err
	}
	if mq.clients[id] != nil {
		_, err = mq.clients[id].Write([]byte(str + "\n"))
	}

	return err
}
