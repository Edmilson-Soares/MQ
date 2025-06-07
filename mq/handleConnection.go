package mq

import (
	"net"

	"github.com/google/uuid"
)

func (mq *MQ) handleConnection(conn net.Conn) {
	id := uuid.New().String()
	defer func() {
		mq.clients[id].Close()
		delete(mq.ips, id)
		delete(mq.clients, id)
	}()
	func() {
		mq.clients[id] = conn
		mq.ips[id] = conn.RemoteAddr().String()
		mq.Send(id, MQData{
			Cmd:     "CNN",
			Topic:   "",
			Payload: id,
		})
	}()

	mq.handleProcess(id, conn)

}
