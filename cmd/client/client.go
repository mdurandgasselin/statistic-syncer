package main

import (
	// "context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	// pb "sync_score/proto"
	sp "sync_score/sport"
	ut "sync_score/utils"

	// "google.golang.org/grpc"
	// "google.golang.org/grpc/credentials/insecure"
	"github.com/streadway/amqp"
)

const (
	defaultName = "world"
)

var (
	addr = flag.String("addr", "localhost:5672", "the address to connect to")
	fileGames = flag.String("fileGames", "client/gamesRecorded.json", "the filepath to a json list, with filepath to recorded games.")
)

func getRabbitMQConnection() *amqp.Connection {
	connection, err := amqp.Dial(fmt.Sprintf("amqp://guest:guest@%s/", *addr))
	if err != nil {
		panic(err)
	}
	return connection
}

func main() {
	flag.Parse()
	log.Println("The filepath trailing", flag.Args())
	queueConnection := getRabbitMQConnection()
	defer queueConnection.Close()

	// Set up a connection to the server.
	var err error
	gamesPath, err := readInputGameFile(*fileGames)
	if err != nil {
		ut.Fatalf("Error while loading files with the recorded games: %s", err)
	}
	ut.Info(gamesPath)
	var namePath string 
	var game sp.Actions
	wg := &sync.WaitGroup{}
	for i, path := range gamesPath {
		namePath = strings.TrimSuffix(filepath.Base(path), ".json")
		game, err = sp.ReadGameFile(path)
		if err != nil {
			ut.Fatalf("Error while loading game: %d", err)
		}
		wg.Add(1)
		
		go sendGame(queueConnection, namePath, game, wg, i)
	}

	wg.Wait()
	ut.Info("The games have all ended")
}

func readInputGameFile(path string) ([]string, error){
	content, err := os.ReadFile(path)
	if err != nil {
		ut.Fatal(err)
	}
	var input []string
	err = json.Unmarshal(content, &input)
	if err != nil {
		ut.Fatal(err)
	}
	return input, nil
}

func sendGame(conn *amqp.Connection, gameName string, game sp.Actions, wg *sync.WaitGroup, threadNumber int) {
	defer wg.Done()
	channel, err := conn.Channel()
	if err != nil {
		panic(err)
	}
	defer channel.Close()
	// declaring queue with its properties over the the channel opened
	_, err = channel.QueueDeclare(
		"LiveGame", // name
		false,     // durable
		false,     // auto delete
		false,     // exclusive
		false,     // no wait
		nil,       // args
	)
	if err != nil {
		panic(err)
	}
	
	ut.Infof("Game %s started: %s ", gameName, threadNumber)

	current_time := int32(0)
	var diff int32
	for _, action := range game {
		diff = action.Minute - current_time		
		time.Sleep(time.Duration(action.Minute - current_time) * 500*  time.Millisecond)
		current_time += diff
		fmt.Println(action)
		body, _ := json.Marshal(action)
		
		// publishing a message
		err = channel.Publish(
			"",        // exchange
			"LiveGame", // key
			false,     // mandatory
			false,     // immediate
			amqp.Publishing{
				ContentType: "text/plain",
				Body:        body,
			},
		)
		if err != nil {
			log.Fatalf("could not publish: %v", err)
		}
	}
}
