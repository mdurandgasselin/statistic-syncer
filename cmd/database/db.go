package main

import (
	"context"
	_ "database/sql"
	_ "fmt"
	"log"
	"net"

	pb "sync_score/proto" // Update with your actual proto package path
	ut "sync_score/utils"

	_ "github.com/mattn/go-sqlite3" // SQLite driver
	"google.golang.org/grpc"
)

type gameEventServer struct {
	pb.UnimplementedGameCenterServer
}

func (s *gameEventServer) SendGameAction(ctx context.Context, event *pb.Action) (*pb.ActionReply, error) {
	// Log the received event
	ut.Debugf("Received event: Game=%s, Team=%s, Player=%s, Description=%s, Time=%d",
		event.GamePoster, event.Team, event.PlayerName, event.Description, event.Minute)

	
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

	
	// Attach the GameCenterService implementation
	pb.RegisterGameCenterServer(grpcServer, &gameEventServer{})

	// Start serving
	log.Println("Starting gRPC server on port 50051...")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}

}

// func main() {
// 	// Create a listener on TCP port

// 	ut.Infoln("Starting db ")
// 	var query string

// 	query = fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
// 			team STRING,
// 			playerName STRING,
// 			description STRING,
// 			minute INTEGER
// 		);`, "sherbrooke_mtl")
// 	_, err = db.Exec(query)
// 	if err != nil {
// 		ut.Fatal(err)
// 	}
// 	fmt.Println("table created")

// 	// add the element
// 	query = fmt.Sprintf(`INSERT INTO %s (team, playerName, description, minute)
// 		VALUES (%s, %s, %s, %d);`, "sherbrooke_mtl", "'sherbrooke'", "'maxime durand'", "'2pts try'", 17)
// 	fmt.Println(query)
// 	_, err = db.Exec(query)
// 	if err != nil {
// 		ut.Fatal(err)
// 	}
// 	fmt.Println("row inséré")

// 	rows, err := db.Query("SELECT team, playerName, description, minute FROM sherbrooke_mtl")
// 	if err != nil {
// 		ut.Fatal(err)
// 	}
// 	defer rows.Close()
// 	ut.Infoln("selection done")

// 	// Iterate through results
// 	for rows.Next() {
// 		var team, name, description string
// 		var minute int
// 		if err := rows.Scan(&team, &name, &description, &minute); err != nil {
// 			ut.Fatal(err)
// 		}
// 		ut.Infof("User: team=%s, Name=%s, description=%s, minute=%d \n", team, name, description, minute)
// 	}

// 	// Return the same event (you could modify or add additional logic)
// }
