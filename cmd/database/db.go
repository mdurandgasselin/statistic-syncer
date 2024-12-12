package main

import (
	"context"
	_ "database/sql"
	"fmt"
	_ "fmt"
	"log"
	"net"

	database "sync_score/cmd/database/db"
	pb "sync_score/proto" // Update with your actual proto package path
	sp "sync_score/sport"
	ut "sync_score/utils"

	"google.golang.org/grpc"
)

type GameEventServer struct {
	pb.UnimplementedGameCenterServer
	db database.DBWrapper
}

func (s *GameEventServer) SendGameAction(ctx context.Context, event *pb.Action) (*pb.ActionReply, error) {
	// Log the received event
	ut.Debugf("Received event: Game=%s, Team=%s, Player=%s, Description=%s, Time=%d",
		event.GamePoster, event.Team, event.PlayerName, event.Description, event.Minute)

	act := sp.Action{
		GamePoster:  event.GamePoster,
		Team:        event.Team,
		PlayerName:  event.PlayerName,
		Description: event.Description,
		Minute:      event.Minute,
	}
	s.db.SendToTables(act)
	// Return the same event (you could modify or add additional logic)
	return &pb.ActionReply{Status: "received"}, nil
}

func (s *GameEventServer) GetGameRecord(ctx context.Context, event *pb.GameTitle) (*pb.Actions, error) {

	spActions, err := s.db.QueryGameHistoric(event.GamePoster)
	if err != nil {
		ut.Debug(err)
		ut.Fatal(err)
	}
	// transform into protobuf messages
	var actions pb.Actions
	for _, act := range spActions {
		actions.Elements = append(actions.Elements, &pb.Action{
			GamePoster:  act.GamePoster,
			Team:        act.Team,
			PlayerName:  act.PlayerName,
			Description: act.Description,
			Minute:      act.Minute})
	}

	return &actions, nil
}

func GetGameEventServer(dbName string) *GameEventServer {
	db := database.NewDBWrapper(dbName)
	t := &GameEventServer{db: db}
	return t
}

func PrintSomething() {
	fmt.Println("Something")
}

func main() {
	// Create a listener on TCP port
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	// Create a gRPC server object
	grpcServer := grpc.NewServer()

	// db := database.NewDBWrapper()

	// Attach the GameCenterService implementation  TODO NewServer
	pb.RegisterGameCenterServer(grpcServer, GetGameEventServer("./games.db"))

	// Start serving
	log.Println("Starting gRPC server on port 50051...")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
