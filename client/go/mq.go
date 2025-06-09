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

type Document map[string]interface{}
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

type JSData struct {
	ID      string            `json:"id"`
	App     map[string]string `json:"app"`
	Env     map[string]string `json:"env"`
	Stripts map[string]string `json:"stripts"`
}

type MQResponse struct {
	Payload string `json:"payload"`
	Error   string `json:"error"`
}
type DbCollection struct {
	name string
	send func(collection, data, type_ string) (string, error)
}

type KV struct {
	bucket string
	send   func(bucket, key, value, type_ string) (string, error)
}
type ScriptJS struct {
	name string
	send func(name, key, value, type_ string) (string, error)
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

	go mq.on()
	_, err = mq.connect(info.User, info.Pass, 1*time.Second)
	if err != nil {
		return nil, err
	}
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

func (mq *MQ) Subscribe(topic string, cb func(msg MQData)) {
	mq.subs[topic] = append(mq.subs[topic], cb)
	mq.Send(MQData{
		Cmd:     "SUB",
		Topic:   topic,
		Payload: "",
	})
}

func (mq *MQ) Publish(topic, Payload string) {
	mq.Send(MQData{
		Cmd:     "PUB",
		Topic:   topic,
		Payload: Payload,
	})
}
func (mq *MQ) connect(username, password string, timeout time.Duration) (string, error) {
	reqId := uuid.New().String()
	mq.chrequest[reqId] = make(chan MQResponse) // cria um canal de string
	mq.Send(MQData{
		Cmd:       "AUTH",
		Topic:     username,
		RequestId: reqId,
		Payload:   password,
	})
	ch, existe := mq.chrequest[reqId]
	if !existe {
		return "", fmt.Errorf("canal %s não existe", username)
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
		return "", fmt.Errorf("timeout de %v expirado no canal %s", timeout, username)
	}
}
func (mq *MQ) Request(topic, Payload string, timeout time.Duration) (string, error) {
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

// ////////////
func (mq *MQ) Script(name string) *ScriptJS {
	return &ScriptJS{
		send: func(name, key, value, type_ string) (string, error) {
			reqId := uuid.New().String()
			mq.chrequest[reqId] = make(chan MQResponse) // cria um canal de string
			tipic := ""
			if name != "" {
				tipic = name
				if key != "" {
					tipic = name + ":" + key
				}
			} else {
				tipic = key
			}

			mq.Send(MQData{
				Cmd:       type_,
				Topic:     tipic,
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
			case <-time.After(2 * time.Second):
				close(ch)
				return "", fmt.Errorf("timeout DbCreateCollection")
			}
		},
		name: name,
	}
}
func (js *ScriptJS) SetEnv(env map[string]string) (string, error) {
	strInput, _ := json.Marshal(env)
	str, err := js.send(js.name, "", string(strInput), "S_ENV")
	if err != nil {
		return "", err
	}
	return str, nil
}
func (js *ScriptJS) SetApp(app map[string]string) (string, error) {
	strInput, _ := json.Marshal(app)
	str, err := js.send(js.name, "", string(strInput), "S_APP")
	if err != nil {
		return "", err
	}
	return str, nil
}
func (js *ScriptJS) SetScript(scripts map[string]string) (string, error) {
	strInput, _ := json.Marshal(scripts)
	str, err := js.send(js.name, "", string(strInput), "S_JS")
	if err != nil {
		return "", err
	}
	return str, nil
}
func (js *ScriptJS) Run() (string, error) {
	str, err := js.send(js.name, "", "", "S_RUN")
	if err != nil {
		return "", err
	}
	return str, nil
}
func (js *ScriptJS) Stop() (string, error) {
	str, err := js.send(js.name, "", "", "S_STOP")
	if err != nil {
		return "", err
	}
	return str, nil
}

func (js *ScriptJS) Register(input JSData) error {
	input.ID = js.name
	strInput, _ := json.Marshal(input)
	_, err := js.send(js.name, "", string(strInput), "S_ADD")
	if err != nil {
		return err
	}
	return nil
}
func (js *ScriptJS) Remove() error {
	_, err := js.send(js.name, "", "", "S_DEL")
	if err != nil {
		return err
	}
	return nil
}

// ///////////////////////////

func (mq *MQ) Kv(bucket string) *KV {
	return &KV{
		send: func(bucket, key, value, type_ string) (string, error) {
			reqId := uuid.New().String()
			mq.chrequest[reqId] = make(chan MQResponse) // cria um canal de string
			tipic := ""
			if bucket != "" {
				tipic = bucket
				if key != "" {
					tipic = bucket + ":" + key
				}
			} else {
				tipic = key
			}

			mq.Send(MQData{
				Cmd:       type_,
				Topic:     tipic,
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
			case <-time.After(2 * time.Second):
				close(ch)
				return "", fmt.Errorf("timeout DbCreateCollection")
			}
		},
		bucket: bucket,
	}
}
func (kv *KV) Set(key, value string) (string, error) {
	str, err := kv.send(kv.bucket, key, value, "SET")
	if err != nil {
		return "", err
	}
	return str, nil
}
func (kv *KV) Get(key string) (string, error) {
	str, err := kv.send(kv.bucket, key, "", "GET")
	if err != nil {
		return "", err
	}
	return str, nil
}
func (kv *KV) Del(key string) (string, error) {
	str, err := kv.send(kv.bucket, key, "", "DEL")
	if err != nil {
		return "", err
	}
	return str, nil
}

func (kv *KV) FilterByKey(filter string) (string, error) {
	str, err := kv.send(kv.bucket, "", filter, "BFK")
	if err != nil {
		return "", err
	}
	return str, nil
}

func (kv *KV) FilterByValue(filter string) (string, error) {
	str, err := kv.send(kv.bucket, "", filter, "BFV")
	if err != nil {
		return "", err
	}
	return str, nil
}
func (kv *KV) CreateBucket() (string, error) {
	str, err := kv.send(kv.bucket, "", "", "BADD")
	if err != nil {
		return "", err
	}
	return str, nil
}
func (kv *KV) DeleteBucket() (string, error) {
	str, err := kv.send(kv.bucket, "", "", "BDEL")
	if err != nil {
		return "", err
	}
	return str, nil
}

// ///////////////////////////////////////
func (mq *MQ) DbCreateCollection(name, indexName string) error {
	reqId := uuid.New().String()
	mq.chrequest[reqId] = make(chan MQResponse) // cria um canal de string
	mq.Send(MQData{
		Cmd:       "DB_CC",
		Topic:     name,
		RequestId: reqId,
		Payload:   indexName,
	})
	ch, existe := mq.chrequest[reqId]
	if !existe {
		return fmt.Errorf("canal %s não existe", "")
	}

	select {
	case res := <-ch:
		close(ch)
		if res.Error != "" {
			return errors.New("Error :" + res.Error)
		}
		return nil
	case <-time.After(2 * time.Second):
		close(ch)
		return fmt.Errorf("timeout DbCreateCollection")
	}
}
func (mq *MQ) DbDeleteCollection(name string) error {
	reqId := uuid.New().String()
	mq.chrequest[reqId] = make(chan MQResponse) // cria um canal de string
	mq.Send(MQData{
		Cmd:       "DB_CC",
		Topic:     name,
		RequestId: reqId,
		Payload:   "",
	})
	ch, existe := mq.chrequest[reqId]
	if !existe {
		return fmt.Errorf("canal %s não existe", "")
	}

	select {
	case res := <-ch:
		close(ch)
		if res.Error != "" {
			return errors.New("Error :" + res.Error)
		}
		return nil
	case <-time.After(2 * time.Second):
		close(ch)
		return fmt.Errorf("timeout de %v expirado no canal", 1)
	}
}

func (mq *MQ) DbCollection(name string) *DbCollection {
	return &DbCollection{
		send: func(collection, data, type_ string) (string, error) {
			reqId := uuid.New().String()
			mq.chrequest[reqId] = make(chan MQResponse) // cria um canal de string
			mq.Send(MQData{
				Cmd:       type_,
				Topic:     collection,
				RequestId: reqId,
				Payload:   data,
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
			case <-time.After(2 * time.Second):
				close(ch)
				return "", fmt.Errorf("timeout DbCreateCollection")
			}
		},
		name: name,
	}
}

func (db *DbCollection) Insert(doc Document) (string, error) {
	strInput, _ := json.Marshal(doc)
	return db.send(db.name, string(strInput), "DB_CI")
}
func (db *DbCollection) FindOne(id string) (*Document, error) {
	str, err := db.send(db.name, string(id), "DB_CG")
	if err != nil {
		return nil, err
	}
	docRes := Document{}
	json.Unmarshal([]byte(str), &docRes)
	return &docRes, nil

}

func (db *DbCollection) Find(doc Document) (*[]Document, error) {
	strInput, _ := json.Marshal(doc)
	str, err := db.send(db.name, string(strInput), "DB_CF")
	if err != nil {
		return nil, err
	}
	docRes := []Document{}
	json.Unmarshal([]byte(str), &docRes)
	return &docRes, nil

}
func (db *DbCollection) FindAll() (*[]Document, error) {
	str, err := db.send(db.name, "", "DB_CL")
	if err != nil {
		return nil, err
	}
	docRes := []Document{}
	json.Unmarshal([]byte(str), &docRes)
	return &docRes, nil

}
func (db *DbCollection) Update(id string, doc Document) (string, error) {
	doc["_id"] = id
	strInput, _ := json.Marshal(doc)
	str, err := db.send(db.name, string(strInput), "DB_CU")
	if err != nil {
		return "", err
	}

	return str, nil

}
func (db *DbCollection) Delete(id string) (string, error) {
	str, err := db.send(db.name, string(id), "DB_CR")
	if err != nil {
		return "", err
	}
	return str, nil
}

// //////////////////////
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
			ch, existe := mq.chrequest[data.RequestId]
			if !existe {
				return
			}

			ch <- MQResponse{
				Payload: data.Payload,
				Error:   data.Error,
			}
		case "OK":
			//fmt.Println(data)
		case "ER_AUH":
			mq.ID = data.Payload
			ch, existe := mq.chrequest[data.RequestId]
			if !existe {
				panic(data.Payload)
				return
			}

			ch <- MQResponse{
				Payload: data.Payload,
				Error:   data.Error,
			}
			mq.Stop()
			panic(data.Payload)
			//fmt.Println(data)
		case "S_ADD", "S_DEL", "S_ENV", "S_JS", "S_RUN", "S_STOP", "S_APP":
			ch, existe := mq.chrequest[data.RequestId]
			if !existe {
				return
			}
			ch <- MQResponse{
				Payload: data.Payload,
				Error:   data.Error,
			}
		case "RES", "SET", "GET", "DEL", "BDEL", "BADD", "BFK", "BFV", "DB_CC", "DB_CD", "DB_CI", "DB_CG", "DB_CR", "DB_CF", "DB_CU", "DB_CL":
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
