package client

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/google/uuid"
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
type MQResponse struct {
	Payload string `json:"payload"`
	Error   string `json:"error"`
}

func jsonToStructResponse(data string) (*MQResponse, error) {
	var res MQResponse
	err := json.Unmarshal([]byte(data), &res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}
func structToJSONResponse(res MQResponse) (string, error) {
	bytes, err := json.Marshal(res)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
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

type MQ struct {
	ID        string
	conn      net.Conn
	subs      map[string][]func(msg MQData)
	chs       map[string]chan string     // cria um canal de string
	chrequest map[string]chan MQResponse // cria um canal de string
	services  map[string]func(msg MQData, replay func(err string, payload string))
}

func Dial(url string) (*MQ, error) {

	info, err := ParseMQURL(url)
	if err != nil {
		return nil, err
	}
	conn, err := net.Dial("tcp", info.Host+":"+info.Port)
	if err != nil {
		return nil, err
	}
	mq := MQ{
		conn:      conn,
		chs:       make(map[string]chan string),
		chrequest: make(map[string]chan MQResponse),
		services:  map[string]func(msg MQData, replay func(err string, payload string)){},
		subs:      make(map[string][]func(msg MQData)),
	}
	mq.Send(MQData{
		Cmd:     "AUTH",
		Topic:   "",
		Payload: info.User + ":" + info.Pass,
	})
	go mq.on()
	return &mq, nil
}

func (mq *MQ) Send(data MQData) error {
	str, err := structToJSON(data)
	if err != nil {
		return err
	}
	_, err = mq.conn.Write([]byte(str + "\n"))
	return err
}

func (mq *MQ) Service(topic string, fn func(msg MQData, replay func(err string, payload string))) {
	mq.services[topic] = fn
	mq.Send(MQData{
		Cmd:     "SER",
		Topic:   topic,
		Payload: "",
	})
}

func (mq *MQ) Sub(topic string, cb func(msg MQData)) {
	mq.subs[topic] = append(mq.subs[topic], cb)
	mq.Send(MQData{
		Cmd:     "SUB",
		Topic:   topic,
		Payload: "",
	})
}

func (mq *MQ) Pub(topic, Payload string) {
	mq.Send(MQData{
		Cmd:     "PUB",
		Topic:   topic,
		Payload: Payload,
	})
}

func (mq *MQ) Req(topic, Payload string, timeout time.Duration) (string, error) {
	reqId := uuid.New().String()
	mq.chrequest[reqId] = make(chan MQResponse) // cria um canal de string
	mq.Send(MQData{
		Cmd:       "REQ",
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

func (mq *MQ) Ping() (string, error) {
	reqId := uuid.New().String()
	mq.chs[reqId] = make(chan string) // cria um canal de string
	mq.Send(MQData{
		Cmd:       "PING",
		Topic:     "",
		RequestId: reqId,
		Payload:   "",
	})
	ch, existe := mq.chs[reqId]
	if !existe {
		return "", fmt.Errorf("canal %s não existe", "")
	}

	select {
	case res := <-ch:
		close(ch)
		return res, nil
	case <-time.After(1 * time.Second):
		close(ch)
		return "", fmt.Errorf("timeout de %v expirado no canal", 1)
	}
}

func (mq *MQ) Get(key string) (string, error) {
	reqId := uuid.New().String()
	mq.chrequest[reqId] = make(chan MQResponse) // cria um canal de string
	mq.Send(MQData{
		Cmd:       "GET",
		Topic:     key,
		RequestId: reqId,
		Payload:   "",
	})
	ch, existe := mq.chrequest[reqId]
	if !existe {
		return "", fmt.Errorf("canal %s não existe", "")
	}

	select {
	case res := <-ch:
		close(ch)
		if res.Error != "" {
			return "", errors.New("Error :" + res.Error)
		}
		return res.Payload, nil
	case <-time.After(1 * time.Second):
		close(ch)
		return "", fmt.Errorf("timeout de %v expirado no canal", 1)
	}
}
func (mq *MQ) Set(key, value string) (string, error) {
	reqId := uuid.New().String()
	mq.chrequest[reqId] = make(chan MQResponse) // cria um canal de string
	mq.Send(MQData{
		Cmd:       "SET",
		Topic:     key,
		RequestId: reqId,
		Payload:   value,
	})
	ch, existe := mq.chrequest[reqId]
	if !existe {
		return "", fmt.Errorf("canal %s não existe", "")
	}

	select {
	case res := <-ch:
		close(ch)
		if res.Error != "" {
			return "", errors.New("Error :" + res.Error)
		}
		return res.Payload, nil
	case <-time.After(1 * time.Second):
		close(ch)
		return "", fmt.Errorf("timeout de %v expirado no canal", 1)
	}
}

func (mq *MQ) Del(key string) (string, error) {
	reqId := uuid.New().String()
	mq.chrequest[reqId] = make(chan MQResponse) // cria um canal de string
	mq.Send(MQData{
		Cmd:       "DEL",
		Topic:     key,
		RequestId: reqId,
		Payload:   "",
	})
	ch, existe := mq.chrequest[reqId]
	if !existe {
		return "", fmt.Errorf("canal %s não existe", "")
	}

	select {
	case res := <-ch:
		close(ch)
		if res.Error != "" {
			return "", errors.New("Error :" + res.Error)
		}
		return res.Payload, nil
	case <-time.After(1 * time.Second):
		close(ch)
		return "", fmt.Errorf("timeout de %v expirado no canal", 1)
	}
}
func (mq *MQ) FilterKey(filter string) (string, error) {
	reqId := uuid.New().String()
	mq.chrequest[reqId] = make(chan MQResponse) // cria um canal de string
	mq.Send(MQData{
		Cmd:       "BFK",
		Topic:     "",
		RequestId: reqId,
		Payload:   "",
	})
	ch, existe := mq.chrequest[reqId]
	if !existe {
		return "", fmt.Errorf("canal %s não existe", "")
	}

	select {
	case res := <-ch:
		close(ch)
		if res.Error != "" {
			return "", errors.New("Error :" + res.Error)
		}
		return res.Payload, nil
	case <-time.After(1 * time.Second):
		close(ch)
		return "", fmt.Errorf("timeout de %v expirado no canal", 1)
	}
}
func (mq *MQ) FilterValue(filter string) (string, error) {
	reqId := uuid.New().String()
	mq.chrequest[reqId] = make(chan MQResponse) // cria um canal de string
	mq.Send(MQData{
		Cmd:       "BFV",
		Topic:     "",
		RequestId: reqId,
		Payload:   "",
	})
	ch, existe := mq.chrequest[reqId]
	if !existe {
		return "", fmt.Errorf("canal %s não existe", "")
	}

	select {
	case res := <-ch:
		close(ch)
		if res.Error != "" {
			return "", errors.New("Error :" + res.Error)
		}
		return res.Payload, nil
	case <-time.After(1 * time.Second):
		close(ch)
		return "", fmt.Errorf("timeout de %v expirado no canal", 1)
	}
}
func (mq *MQ) BFilterKey(bucket, filter string) (string, error) {
	reqId := uuid.New().String()
	mq.chrequest[reqId] = make(chan MQResponse) // cria um canal de string
	mq.Send(MQData{
		Cmd:       "BFK",
		Topic:     bucket,
		RequestId: reqId,
		Payload:   "",
	})
	ch, existe := mq.chrequest[reqId]
	if !existe {
		return "", fmt.Errorf("canal %s não existe", "")
	}

	select {
	case res := <-ch:
		close(ch)
		if res.Error != "" {
			return "", errors.New("Error :" + res.Error)
		}
		return res.Payload, nil
	case <-time.After(1 * time.Second):
		close(ch)
		return "", fmt.Errorf("timeout de %v expirado no canal", 1)
	}
}
func (mq *MQ) BFilterValue(bucket, filter string) (string, error) {
	reqId := uuid.New().String()
	mq.chrequest[reqId] = make(chan MQResponse) // cria um canal de string
	mq.Send(MQData{
		Cmd:       "BFV",
		Topic:     bucket,
		RequestId: reqId,
		Payload:   "",
	})
	ch, existe := mq.chrequest[reqId]
	if !existe {
		return "", fmt.Errorf("canal %s não existe", "")
	}

	select {
	case res := <-ch:
		close(ch)
		if res.Error != "" {
			return "", errors.New("Error :" + res.Error)
		}
		return res.Payload, nil
	case <-time.After(1 * time.Second):
		close(ch)
		return "", fmt.Errorf("timeout de %v expirado no canal", 1)
	}
}
func (mq *MQ) BGet(bucket, key string) (string, error) {
	reqId := uuid.New().String()
	mq.chrequest[reqId] = make(chan MQResponse) // cria um canal de string
	mq.Send(MQData{
		Cmd:       "GET",
		Topic:     bucket + ":" + key,
		RequestId: reqId,
		Payload:   "",
	})
	ch, existe := mq.chrequest[reqId]
	if !existe {
		return "", fmt.Errorf("canal %s não existe", "")
	}

	select {
	case res := <-ch:
		close(ch)
		if res.Error != "" {
			return "", errors.New("Error :" + res.Error)
		}
		return res.Payload, nil
	case <-time.After(1 * time.Second):
		close(ch)
		return "", fmt.Errorf("timeout de %v expirado no canal", 1)
	}
}
func (mq *MQ) BSet(bucket, key, value string) (string, error) {
	reqId := uuid.New().String()
	mq.chrequest[reqId] = make(chan MQResponse) // cria um canal de string
	mq.Send(MQData{
		Cmd:       "SET",
		Topic:     bucket + ":" + key,
		RequestId: reqId,
		Payload:   value,
	})
	ch, existe := mq.chrequest[reqId]
	if !existe {
		return "", fmt.Errorf("canal %s não existe", "")
	}

	select {
	case res := <-ch:
		close(ch)
		if res.Error != "" {
			return "", errors.New("Error :" + res.Error)
		}
		return res.Payload, nil
	case <-time.After(1 * time.Second):
		close(ch)
		return "", fmt.Errorf("timeout de %v expirado no canal", 1)
	}
}

func (mq *MQ) BDel(bucket, key string) (string, error) {
	reqId := uuid.New().String()
	mq.chrequest[reqId] = make(chan MQResponse) // cria um canal de string
	mq.Send(MQData{
		Cmd:       "DEL",
		Topic:     bucket + ":" + key,
		RequestId: reqId,
		Payload:   "",
	})
	ch, existe := mq.chrequest[reqId]
	if !existe {
		return "", fmt.Errorf("canal %s não existe", "")
	}

	select {
	case res := <-ch:
		close(ch)
		if res.Error != "" {
			return "", errors.New("Error :" + res.Error)
		}
		return res.Payload, nil
	case <-time.After(1 * time.Second):
		close(ch)
		return "", fmt.Errorf("timeout de %v expirado no canal", 1)
	}
}

func (mq *MQ) DeleteBucket(name string) (string, error) {
	reqId := uuid.New().String()
	mq.chrequest[reqId] = make(chan MQResponse) // cria um canal de string
	mq.Send(MQData{
		Cmd:       "BDEL",
		Topic:     name,
		RequestId: reqId,
		Payload:   "",
	})
	ch, existe := mq.chrequest[reqId]
	if !existe {
		return "", fmt.Errorf("canal %s não existe", "")
	}

	select {
	case res := <-ch:
		close(ch)
		if res.Error != "" {
			return "", errors.New("Error :" + res.Error)
		}
		return res.Payload, nil
	case <-time.After(1 * time.Second):
		close(ch)
		return "", fmt.Errorf("timeout de %v expirado no canal", 1)
	}
}

func (mq *MQ) CreateBucket(name string) (string, error) {
	reqId := uuid.New().String()
	mq.chrequest[reqId] = make(chan MQResponse) // cria um canal de string
	mq.Send(MQData{
		Cmd:       "BADD",
		Topic:     name,
		RequestId: reqId,
		Payload:   "",
	})
	ch, existe := mq.chrequest[reqId]
	if !existe {
		return "", fmt.Errorf("canal %s não existe", "")
	}

	select {
	case res := <-ch:
		close(ch)
		if res.Error != "" {
			return "", errors.New("Error :" + res.Error)
		}
		return res.Payload, nil
	case <-time.After(1 * time.Second):
		close(ch)
		return "", fmt.Errorf("timeout de %v expirado no canal", 1)
	}
}
func (mq *MQ) on() {
	reader := bufio.NewReader(mq.conn)
	for {
		// Lê a mensagem do cliente até encontrar uma nova linha
		str, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("Erro ao ler: %s\n", err.Error())
			return
		}

		data, err := jsonToStruct(str)
		if err != nil {
			fmt.Printf("Erro ao ler: %s\n", err.Error())
			return
		}
		switch data.Cmd {
		case "CNN":
			mq.ID = data.Payload
			fmt.Println("connected")
			//fmt.Println(data)
		case "OK":
			//fmt.Println(data)
		case "ER_AUH":
			mq.Stop()
			panic(data.Payload)
			//fmt.Println(data)
		case "RES", "SET", "GET", "DEL", "BDEL", "BADD", "BFK", "BFV":
			ch, existe := mq.chrequest[data.RequestId]
			if !existe {
				return
			}

			ch <- MQResponse{
				Payload: data.Payload,
				Error:   data.Error,
			}

		case "PONG":
			ch, existe := mq.chs[data.RequestId]
			if !existe {
				return
			}
			ch <- data.Payload
		case "REQ":
			if mq.services[data.Topic] != nil {
				mq.services[data.Topic](*data, func(err string, payload string) {
					mq.Send(MQData{
						Cmd:       "RES",
						RequestId: data.RequestId,
						ReplayId:  data.ReplayId,
						Topic:     data.Topic,
						Error:     err,
						Payload:   payload,
					})
				})
			}
		case "PUB":
			topic := data.Topic
			if strings.Contains(data.Regtopic, ".*.") {
				topic = data.Regtopic
			}
			for _, sub := range mq.subs[topic] {
				go sub(*data)
			}
		}

	}
}

func (mq *MQ) Stop() error {
	return mq.conn.Close()
}
