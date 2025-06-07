package mq

import (
	"bufio"
	"fmt"
	"net"
)

func (mq *MQ) handleProcess(id string, conn net.Conn) {
	reader := bufio.NewReader(conn)
	for {
		// Lê a mensagem do cliente até encontrar uma nova linha
		str, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("Erro ao ler: %s\n", err.Error())
			break
		}

		data, err := jsonToStruct(str)
		if err != nil {
			fmt.Printf(": %s\n", err.Error())
			return
		}

		switch data.Cmd {
		case "SUB":
			mq.handleSub(id, *data)
		case "SER":
			mq.handleService(id, *data)
		case "RES":
			mq.handleRes(id, *data)
		case "PUB":
			go mq.handlePub(*data)
		case "REQ":
			mq.handleReq(id, *data)
		case "PING":
			mq.Send(id, MQData{
				Cmd:       "PONG",
				ReplayId:  id,
				RequestId: data.RequestId,
				Topic:     data.Topic,
				Payload:   "PONG",
			})

		case "SET":
			mq.handleSet(id, *data)

		case "GET":
			mq.handleGet(id, *data)

		case "DEL":
			mq.handleDel(id, *data)
		case "BDEL":
			mq.handleBDel(id, *data)
		case "BADD":
			mq.handleBAdd(id, *data)

		case "BFV":
			mq.handleBFilterVal(id, *data)
		case "BFK":
			mq.handleBFilterKey(id, *data)

		}
	}
}
