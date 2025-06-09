package server

import (
	"net"

	"github.com/google/uuid"
)

func (mq *MQ) handleConnection(conn net.Conn, reqId string) {
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
			Cmd:       "CNN",
			Topic:     "",
			RequestId: reqId,
			Payload:   id,
		})
	}()

	mq.handleProcess(id, conn)

}
