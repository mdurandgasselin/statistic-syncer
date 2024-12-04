package main

import (
	"context"
	_ "database/sql"
	_ "fmt"
	"log"
	"net"

	pb "sync_score/proto" // Update with your actual proto package path
	sp "sync_score/sport"
	ut "sync_score/utils"
	database "sync_score/cmd/database/db"

	"google.golang.org/grpc"
)

type gameEventServer struct {
	pb.UnimplementedGameCenterServer
	db database.DBWrapper
}

func (s *gameEventServer) SendGameAction(ctx context.Context, event *pb.Action) (*pb.ActionReply, error) {
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

func main() {
	// Create a listener on TCP port
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	// Create a gRPC server object
	grpcServer := grpc.NewServer()

	db := database.NewDBWrapper()

	// Attach the GameCenterService implementation
	pb.RegisterGameCenterServer(grpcServer, &gameEventServer{db: db})

	// Start serving
	log.Println("Starting gRPC server on port 50051...")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
