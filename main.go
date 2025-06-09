package main

import (
	"fmt"
	"log"
	"mq/cmd/server"
	"mq/utils"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// Ouve na porta 4051 em todas as interfaces de rede

	config, err := utils.GetConfig()
	if err != nil {
		fmt.Println(err)
		return
	}
	mq := server.NewMQ(config.MQ)
	go mq.Subscribe("test.*.test", func(data server.MQData) {
		fmt.Println(data.Topic, data.Payload)
	})

	mq.Service("test", func(data server.MQData, replay func(err string, data string)) {
		fmt.Println(data.Topic, data.Payload)
		replay("", "okfffffff")
	})

	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				mq.Publish("test.t555.test", "dddddddddddddddddd")
				str, err := mq.Request("testdd", "dddd666ddd", 4*time.Second)
				if err != nil {
					fmt.Println(err)
				}
				fmt.Println(str)
				// Coloque aqui a tarefa que você deseja executar
			}
		}
	}()

	go mq.Start()
	exit := make(chan os.Signal, 1)
	signal.Notify(exit, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	<-exit
	log.Println("Sinal de encerramento recebido, fechando o servidor...")

}
