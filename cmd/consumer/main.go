package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/IamP5/go-demo/internal/order/infra/database"
	"github.com/IamP5/go-demo/internal/order/usecase"
	"github.com/IamP5/go-demo/pkg/rabbitmq"
	_ "github.com/mattn/go-sqlite3"
	amqp "github.com/rabbitmq/amqp091-go"
	"time"
)

func main() {
	db, err := sql.Open("sqlite3", "./orders.db")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	repository := database.NewOrderRepository(db)
	uc := usecase.CalculateFinalPriceUseCase{OrderRepository: repository}

	ch, err := rabbitmq.OpenChannel()
	if err != nil {
		panic(err)
	}
	defer ch.Close()

	out := make(chan amqp.Delivery) //channel
	go rabbitmq.Consume(ch, out)    // multithreading

	for msg := range out {
		var inputDTO usecase.OrderInputDTO
		err := json.Unmarshal(msg.Body, &inputDTO)
		if err != nil {
			panic(err)
		}

		outputDTO, err := uc.Execute(inputDTO)
		if err != nil {
			panic(err)
		}

		msg.Ack(false)
		fmt.Println(outputDTO)
		time.Sleep(500 * time.Millisecond)
	}
}
