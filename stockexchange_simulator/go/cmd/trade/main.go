package main

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/andrepreira/homebroker_app/stockexchange_simulator/go/internal/infra/kafka"
	"github.com/andrepreira/homebroker_app/stockexchange_simulator/go/internal/market/dto"
	"github.com/andrepreira/homebroker_app/stockexchange_simulator/go/internal/market/entity"
	"github.com/andrepreira/homebroker_app/stockexchange_simulator/go/internal/market/transformer"
	ckafka "github.com/confluentinc/confluent-kafka-go/kafka"
)

func main() {
	ordersIn := make(chan *entity.Order)
	ordersOut := make(chan *entity.Order)
	wg := &sync.WaitGroup{}
	defer wg.Wait()

	kafkaMsgChan := make(chan *ckafka.Message)
	configMap := &ckafka.ConfigMap{
		"bootstrap.servers": "host.docker.internal:9094",
		"group.id":          "myGroup",
		"auto.offset.reset": "latest",
	}
	producer := kafka.NewKafkaProducer(configMap)
	kafka := kafka.NewConsumer(configMap, []string{"input"})

	go kafka.Consume(kafkaMsgChan) // T2

	// receive kafkaMsgChan and send to ordersIn and return ordersOut
	book := entity.NewBook(ordersIn, ordersOut, wg)

	go book.Trade() // T3

	go func() {
		for msg := range kafkaMsgChan {
			wg.Add(1)
			fmt.Println(string(msg.Value))
			tradeInput := dto.TradeInput{}
			err := json.Unmarshal(msg.Value, &tradeInput)
			if err != nil {
				panic(err)
			}
			order := transformer.TransformInput(tradeInput)
			ordersIn <- order
		}
	}()

	for res := range ordersOut {
		output := transformer.TransformOutput(res)
		outputJson, err := json.MarshalIndent(output, "", "  ")
		fmt.Println(string(outputJson))
		if err != nil {
			fmt.Println(err)
		}

		producer.Publish(outputJson, []byte("order"), "output")
	}
}
