package server

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
			//KV
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
			//NoSQL
		case "BFV":
			mq.handleBFilterVal(id, *data)
		case "BFK":
			mq.handleBFilterKey(id, *data)
		case "DB_CC":
			mq.handledbCreateCollection(id, *data)
		case "DB_CD":
			mq.handledbDeleteCollection(id, *data)
		case "DB_CI":
			mq.handledbInsertCollection(id, *data)
		case "DB_CG":
			mq.handledbFindOneCollection(id, *data)
		case "DB_CR":
			mq.handledbRemoveCollection(id, *data)
		case "DB_CF":
			mq.handledbFilterCollection(id, *data)
		case "DB_CU":
			mq.handledbUpdateCollection(id, *data)
		case "DB_CL":
			mq.handledbListCollection(id, *data)

			//////////Script
		case "S_ADD":
			mq.handleScriptJsAdd(id, *data)
		case "S_DEL":
			mq.handleScriptJsDel(id, *data)
		case "S_ENV":
			mq.handleScriptJsAdd(id, *data)
		case "S_JS":
			mq.handleScriptJsAdd(id, *data)
		case "S_RUN":
			mq.handleScriptJsAdd(id, *data)
		case "S_STOP":
			mq.handleScriptJsAdd(id, *data)
		}
	}
}
