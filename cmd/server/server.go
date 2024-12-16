package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"sync"
	"time"
	"strings"

	pb "statistic-syncer/proto"
	sp "statistic-syncer/sport"
	ut "statistic-syncer/utils"

	"github.com/streadway/amqp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	// pb "statistic-syncer/proto"
	// "google.golang.org/grpc"
	// "google.golang.org/grpc/credentials/insecure"
)

var (
	rabbitQueuePort = flag.Int("rabbitQueuePort", 5672, "The rabbitMQ port")
)

func main() {
	flag.Parse()

	// lis, err := net.Listen("tcp", ":50051")
	// if err != nil {
	// 	log.Fatalf("Failed to listen: %v", err)
	// }

	// // Create a gRPC server object
	// grpcServer := grpc.NewServer()

	// // Attach the GameCenterService implementation  TODO NewServer
	// pb.RegisterGameCenterServer(grpcServer, ...)


	cacheGameRecorded := NewCacheGameRecorded(5 * time.Second)
	cacheGameRecorded.start()

	// Server MQTT
	server, err := NewQueueServer(cacheGameRecorded)
	if err != nil {
		ut.Fatalf("Failed to create server: %v", err)
	}
	defer server.Close()

	if err := server.Start(); err != nil {
		ut.Fatalf("Server error: %v", err)
	}
}

type CacheGameRecorded struct {
	games map[string]sp.ScoreRecord
	mu    sync.Mutex
	ttl   time.Duration
}

func NewCacheGameRecorded(ttl time.Duration) *CacheGameRecorded {
	return &CacheGameRecorded{
		games: make(map[string]sp.ScoreRecord),
		ttl:   ttl,
	}
}

func (c *CacheGameRecorded) updateCache(action sp.Action) {
	var increaseScore int32
	switch action.Description {
	case "free throw succes":
		increaseScore = 1
	case "2pts succes":
		increaseScore = 2
	case "3pts succes":
		increaseScore = 3
	default:
		return
	}
	c.mu.Lock()
	{
		var sRecord sp.ScoreRecord
		sRecord, ok := c.games[action.GamePoster]
		if !ok {
			team1, team2, _ := strings.Cut(action.GamePoster, "_")  // team1 = "Boston", team2 = "Knicks"
			c.games[action.GamePoster] = sp.ScoreRecord{
				GameName: action.GamePoster,
				TeamA:    team1,
				TeamB:    team2,
				ScoreA:   0,
				ScoreB:   0,
			}
			sRecord = c.games[action.GamePoster]
		}
		if action.Team == sRecord.TeamA {
			sRecord.ScoreA += increaseScore
		} else {
			sRecord.ScoreB += increaseScore
		}
		
		c.games[action.GamePoster] = sRecord
	}
	c.mu.Unlock()
}

func (c *CacheGameRecorded) getScore(gamePoster string) sp.ScoreRecord {
	c.mu.Lock()
	defer c.mu.Unlock()
	record, ok := c.games[gamePoster]
	if !ok {
		return sp.ScoreRecord{}
	}
	record.Reset()
	return record
}

func (c *CacheGameRecorded) clearCacheIfExpired() {
	c.mu.Lock()
	defer c.mu.Unlock()
	for k, v := range c.games {
		if time.Since(v.LastRead) > c.ttl {
			delete(c.games, k)
		} else if time.Since(v.LastRead) < 0 {
			v.Reset()
			c.games[k] = v
		}
	}
}

func (c *CacheGameRecorded) start() {
	go func() {
		for range time.Tick(c.ttl) {
			c.clearCacheIfExpired()
		}
	}()
}

type QueueServer struct {
	// Unexported field
	grpcDBClient pb.GameCenterDatabaseClient
	amqpConn     *amqp.Connection
	amqpChan     *amqp.Channel
	cacheGameRecorded *CacheGameRecorded
}

func NewQueueServer(cacheGameRecorded *CacheGameRecorded) (*QueueServer, error) {
	// Setup gRPC connection
	conn, err := grpc.NewClient("localhost:50051",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to gRPC server: %v", err)
	}

	// Setup AMQP connection
	amqpConn, err := amqp.Dial(fmt.Sprintf("amqp://guest:guest@localhost:%d/", *rabbitQueuePort))
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %v", err)
	}

	channel, err := amqpConn.Channel()
	if err != nil {
		amqpConn.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %v", err)
	}

	return &QueueServer{
		grpcDBClient: pb.NewGameCenterDatabaseClient(conn),
		amqpConn:     amqpConn,
		amqpChan:     channel,
		cacheGameRecorded: cacheGameRecorded,
	}, nil
}

func (s *QueueServer) Close() {
	if s.amqpChan != nil {
		s.amqpChan.Close()
	}
	if s.amqpConn != nil {
		s.amqpConn.Close()
	}
}

func (s *QueueServer) initQueue(queueName string) (<-chan amqp.Delivery, error) {
	// Create queue if it doesn't exist
	_, err := s.amqpChan.QueueDeclare(
		queueName,
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to declare queue: %v", err)
	}

	msgs, err := s.amqpChan.Consume(
		queueName,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to register consumer: %v", err)
	}
	return msgs, nil
}

func (s *QueueServer) Start() error {
	msgs, err := s.initQueue("LiveGame")
	if err != nil {
		return err
	}

	ut.Info("Successfully connected to RabbitMQ instance")
	ut.Info("Starting to consume messages...")

	var action sp.Action
	for msg := range msgs {
		if err := json.Unmarshal(msg.Body, &action); err != nil {
			ut.Fatalf("Failed to unmarshal message: %v", err)
			continue
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
		s.cacheGameRecorded.updateCache(action)
		response, err := s.grpcDBClient.SendGameAction(ctx, event)
		cancel()
		if err != nil {
			ut.Fatalf("Error sending event: %v", err)
			continue
		}
		ut.Debug(response.Status)
	}

	return nil
}
