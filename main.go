package main

import (
	"fmt"
	"tcp/mq"
	"time"
)

func main() {
	// Ouve na porta 8080 em todas as interfaces de rede

	mqSever := mq.NewMQ()
	go mqSever.Sub("test.*.test", func(data mq.MQData) {
		fmt.Println(data.Topic, data.Payload)
	})

	mqSever.Service("test", func(data mq.MQData, replay func(err string, data string)) {
		fmt.Println(data.Topic, data.Payload)
		replay("", "okfffffff")
	})

	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				mqSever.Pub("test.t555.test", "dddddddddddddddddd")
				str, err := mqSever.Req("testdd", "dddd666ddd", 4*time.Second)
				if err != nil {
					fmt.Println(err)
				}
				fmt.Println(str)
				// Coloque aqui a tarefa que você deseja executar
			}
		}
	}()

	mqSever.Start(":8080")

}
