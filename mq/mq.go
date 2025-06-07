package mq

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
)

type MQData struct {
	Cmd       string `json:"cmd"`
	Topic     string `json:"topic"`
	Payload   string `json:"payload"`
	RequestId string `json:"requestId"`
	ReplayId  string `json:"replayId"`
	Regtopic  string `json:"regtopic"`
	Error     string `json:"error"`
	FromId    string `json:"fromId"`
}

func jsonToStruct(data string) (*MQData, error) {
	var mq MQData
	err := json.Unmarshal([]byte(data), &mq)
	if err != nil {
		return nil, err
	}
	return &mq, nil
}
func structToJSON(mq MQData) (string, error) {
	bytes, err := json.Marshal(mq)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

type MQResponse struct {
	Payload string `json:"payload"`
	Error   string `json:"error"`
}

type MQ struct {
	clients     map[string]net.Conn
	ips         map[string]string
	services    map[string]string
	subs        map[string][]string
	KV          *MQKV
	subself     map[string][]func(data MQData)
	chs         map[string]chan string     // cria um canal de string
	chrequest   map[string]chan MQResponse // cria um canal de string
	serviceself map[string]func(data MQData, replay func(err string, payload string))
}

func (mq *MQ) Start(url string) {

	listener, err := net.Listen("tcp", url)
	if err != nil {
		log.Fatalf("Erro ao iniciar o servidor: %s", err.Error())
	}
	defer listener.Close()

	fmt.Println("Servidor TCP iniciado e ouvindo na porta 8080...")

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("Erro ao aceitar conexão: %s", err.Error())
			continue
		}
		err = mq.handleAuth(conn)
		if err != nil {
			mq.send(conn, MQData{
				Cmd:       "ER_AUH",
				RequestId: "",
				Payload:   err.Error(),
			})
			conn.Close()
		} else {
			go mq.handleConnection(conn)
		}

	}
}

func NewMQ() *MQ {

	mq := MQ{
		clients:     map[string]net.Conn{},
		subself:     make(map[string][]func(data MQData)),
		chs:         make(map[string]chan string),
		KV:          NewMQKV("store.db"),
		ips:         make(map[string]string),
		services:    make(map[string]string),
		subs:        make(map[string][]string),
		chrequest:   make(map[string]chan MQResponse),
		serviceself: make(map[string]func(data MQData, replay func(err string, payload string))),
	}

	return &mq
}
