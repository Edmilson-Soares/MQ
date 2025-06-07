package mq

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"strings"
)

func (mq *MQ) handleAuth(conn net.Conn) error {

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
			return err
		}

		switch data.Cmd {
		case "AUTH":
			fmt.Println(data)
			auth := strings.Split(data.Payload, ":")

			if auth[0] == "user" && auth[1] == "pass" {
				return nil
			}

			return errors.New("Errr")

		}

	}

	return nil
}
