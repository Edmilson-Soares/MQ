package server

import (
	"bufio"
	"errors"
	"fmt"
	"net"
)

func (mq *MQ) handleAuth(conn net.Conn) (string, error) {
	reqId := ""
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
			return reqId, err
		}
		reqId = data.RequestId
		switch data.Cmd {
		case "AUTH":
			user := mq.auth[data.Topic]
			if user != data.Payload {
				return reqId, errors.New("Invalid auth")
			}
			return reqId, nil
		}

	}

	return reqId, nil
}
