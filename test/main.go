package main

import (
	"fmt"
	"log"
	"tcp/client"
	"time"
	// Importamos o pacote time para usar um pequeno delay, se necessário
)

func main() {
	// Conecta-se ao servidor TCP na porta 8080 do localhost
	mq, err := client.Dial("mq://user:pass@localhost:8080")
	if err != nil {
		log.Fatalf("Erro ao conectar ao servidor: %s", err.Error())
	}
	defer mq.Stop()

	fmt.Println("Conectado ao servidor TCP em localhost:8080")
	pong, err := mq.Ping()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("dd", pong)
	mq.Service("testdd", func(msg client.MQData, replay func(err string, payload string)) {
		fmt.Println(msg.Payload, "dddd")
		replay("ddddd", "ddddd")
	})

	mq.Sub("test.*.test", func(msg client.MQData) {
		fmt.Println(msg.Payload, msg.Topic, "dsd")
	})

	mq.Pub("test.te.test", "ddddddddddd")
	//mq.Pub("test.t444.test", "ddddddddddd")
	res, err := mq.Req("test", "dddddd", 10*time.Second)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(res, "sttttt")
	mq.CreateBucket("starts")
	//_, err = mq.BSet("starts", "test", "yyyyyyyyyyyyyyyyyyyyyyyyyyyyy")
	res, _ = mq.BGet("starts", "test")
	fmt.Println(res, "sttttt")

	f, _ := mq.FilterKey("t")
	fmt.Println(f, "sttttt")
	select {}

}
