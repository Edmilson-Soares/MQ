package main

import (
	"fmt"
	"log"
	client "mq/client/go"
	// Importamos o pacote time para usar um pequeno delay, se necessário
)

func main() {
	// Conecta-se ao servidor TCP na porta 8080 do localhost
	mq, err := client.Dial("mq://root:fffffffffffffffffff@localhost:4051")
	if err != nil {
		log.Fatalf("Erro ao conectar ao servidor: %s", err.Error())
	}
	defer mq.Stop()

	fmt.Println("Conectado ao servidor TCP em localhost:4051")
	pong, err := mq.Ping()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("dd", pong)

	mq.Script("nodejs").Register(client.JSData{
		ID:  "",
		App: map[string]string{"id": "dddd"},
		Env: map[string]string{"appId": "dddddd"},
		Stripts: map[string]string{
			"index": "console.log('Hello, World!');",
		},
	})

	/*

			mq.Service("testdd", func(msg client.MQData, replay func(err string, payload string)) {
			fmt.Println(msg.Payload, "dddd")
			replay("ddddd", "ddddd")
		})

		mq.Subscribe("test.*.test", func(msg client.MQData) {
			fmt.Println(msg.Payload, msg.Topic, "dsd")
		})

		mq.Publish("test.te.test", "ddddddddddd")
		//mq.Pub("test.t444.test", "ddddddddddd")
		res, err := mq.Request("test", "dddddd", 10*time.Second)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(res, "sttttt")
		mq.Kv("starts").CreateBucket()
		_, err = mq.Kv("starts").Set("test", "yyyyyyyyyyyyyyyyyyyyyyyyyyyyy")
		res, _ = mq.Kv("starts").Get("test")
		fmt.Println(res, "sttttt")

		f, _ := mq.Kv("starts").FilterByKey("tes")
		fmt.Println(f, "sttttt>>")


		err = mq.DbCreateCollection("users", "name")
			if err != nil {
				fmt.Println(err)
			}
			mq.DbCollection("users").Insert(client.Document{
				"nome":  "João",
				"idade": 25,
			})
			user, err := mq.DbCollection("users").FindOne("1749476853203693700")
			fmt.Println(user, err)
		/////////////
				users, err := mq.DbCollection("users").Find(client.Document{
				"nome": "João",
			})
			fmt.Println(users, err)

			/////////////
				re, err := mq.DbCollection("users").Update("1749476853203693700", client.Document{
				"nome": "João Test",
			})
			fmt.Println(re, err)
			//////////////////
				re, err := mq.DbCollection("users").Delete("1749476853203693700")
		fmt.Println(re, err)

			users, err := mq.DbCollection("users").FindAll()
		fmt.Println(users, err)

			err = mq.DbDeleteCollection("users")
		if err != nil {
			fmt.Println(err)
		}
	*/

	select {}

}
