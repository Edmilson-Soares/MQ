package server

import (
	"encoding/json"
	"fmt"
	"mq/cmd/db"
	"mq/utils"
	"net"
	"strconv"
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
	auth        map[string]string
	config      utils.MQConfig
	subs        map[string][]string
	DB          *db.NoSQL
	subself     map[string][]func(data MQData)
	chs         map[string]chan string     // cria um canal de string
	chrequest   map[string]chan MQResponse // cria um canal de string
	serviceself map[string]func(data MQData, replay func(err string, payload string))
}

func (mq *MQ) Start() error {

	listener, err := net.Listen("tcp", mq.config.Broker+":"+strconv.Itoa(mq.config.Port))
	if err != nil {
		return err
	}
	defer listener.Close()

	fmt.Println("Servidor TCP iniciado e ouvindo na " + mq.config.Broker + ":" + strconv.Itoa(mq.config.Port))

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("Erro ao aceitar conexão: %s", err.Error())
			continue
		}
		reqId, err := mq.handleAuth(conn)
		if err != nil {
			mq.send(conn, MQData{
				Cmd:       "ER_AUH",
				RequestId: reqId,
				Payload:   err.Error(),
			})
			conn.Close()
		} else {
			go mq.handleConnection(conn, reqId)
		}

	}
}

func NewMQ(config utils.MQConfig) *MQ {
	dbNoSQL, _ := db.New(config.FileKV)
	mq := MQ{
		clients:     map[string]net.Conn{},
		config:      config,
		auth:        map[string]string{config.Username: config.Password},
		subself:     make(map[string][]func(data MQData)),
		chs:         make(map[string]chan string),
		DB:          dbNoSQL,
		ips:         make(map[string]string),
		services:    make(map[string]string),
		subs:        make(map[string][]string),
		chrequest:   make(map[string]chan MQResponse),
		serviceself: make(map[string]func(data MQData, replay func(err string, payload string))),
	}

	return &mq
}
