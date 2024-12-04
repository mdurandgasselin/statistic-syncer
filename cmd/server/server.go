package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"time"

	pb "sync_score/proto"
	sp "sync_score/sport"
	ut "sync_score/utils"

	"github.com/streadway/amqp"
	"google.golang.org/grpc"

	"google.golang.org/grpc/credentials/insecure"
)

var (
	rabbitQueuePort = flag.Int("rabbitQueuePort", 5672, "The rabbitMQ port")
)

func main() {

	conn, err := grpc.NewClient("localhost:50051",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		ut.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	clientDBRPC := pb.NewGameCenterClient(conn)

	connection, err := amqp.Dial(fmt.Sprintf("amqp://guest:guest@localhost:%d/", *rabbitQueuePort))
	if err != nil {
		panic(err)
	}
	defer connection.Close()
	ut.Info("Successfully connected to RabbitMQ instance")

	channel, err := connection.Channel()
	if err != nil {
		panic(err)
	}
	defer channel.Close()
	msgs := initQueue(channel, "LiveGame")

	var action sp.Action

	for msg := range msgs {
		if err := json.Unmarshal(msg.Body, &action); err != nil {
			panic(err)
		}
		ut.Debugf("Game: %s \n \t Team: %s \n \t name of the player: %s \n \t description: %s \n \t time in minute: %d \n",
			action.GamePoster, action.Team, action.PlayerName, action.Description, action.Minute)

		event := &pb.Action{
			GamePoster:  action.GamePoster,
			Team:        action.Team,
			PlayerName:  action.PlayerName,
			Description: action.Description,
			Minute:      action.Minute,
		}

		// Send event to server
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		response, err := clientDBRPC.SendGameAction(ctx, event)
		if err != nil {
			ut.Fatalf("Error sending event: %v", err)
		}
		ut.Debug(response.Status)

	}
}

func initQueue(channel *amqp.Channel, queueName string) <-chan amqp.Delivery {
	// create if does not exist
	_, err := channel.QueueDeclare(
		queueName,
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		panic(err)
	}

	msgs, err := channel.Consume(
		queueName,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		panic(err)
	}
	return msgs
}
