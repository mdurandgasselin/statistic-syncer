package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"strings"
	"testing"

	pb "sync_score/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

func newServer() (pb.GameCenterClient, func()) {
	lis := bufconn.Listen(1024 * 1024)

	srv := grpc.NewServer()

	pb.RegisterGameCenterServer(
		srv,
		GetGameEventServer("./test_games3.db"),
	)

	go func() {
		if err := srv.Serve(lis); err != nil {
			log.Fatalf("srv.Serve %v", err)
		}
	}()

	dialer := func(context.Context, string) (net.Conn, error) {
		return lis.Dial()
	}

	conn, err := grpc.NewClient(
		"passthrough://",
		grpc.WithContextDialer(dialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("grpc.DialContext %v", err)
	}

	closer := func() {
		lis.Close()
		srv.Stop()
		conn.Close()
	}

	client := pb.NewGameCenterClient(conn)
	return client, closer
}

func TestGameCenterServer_GetGameRecord(t *testing.T) {
	client, closer := newServer()
	defer closer()
	res, err := client.GetGameRecord(context.Background(), &pb.GameTitle{GamePoster: "testingGame"})
	if err != nil {
		t.Fatalf("client.GetUser %v", err)
	}
	// Parse the rows into a slice of Action structs
	ref := referenceGameRecorded()

	for i, elemt := range res.Elements {
		if elemt.GamePoster != ref.Elements[i].GamePoster {
			t.Fatalf("Unexpected value for %s at index %d: %s should be %s", "GamePoster", i, elemt.GamePoster, ref.Elements[i].GamePoster)
		}
		if elemt.Team != ref.Elements[i].Team {
			t.Fatalf("Unexpected value for %s at index %d: %s should be %s", "Team", i, elemt.Team, ref.Elements[i].Team)
		}
		if elemt.PlayerName != ref.Elements[i].PlayerName {
			t.Fatalf("Unexpected value for %s at index %d: %s should be %s", "PlayerName", i, elemt.PlayerName, ref.Elements[i].PlayerName)
		}
		if elemt.Description != ref.Elements[i].Description {
			t.Fatalf("Unexpected value for %s at index %d: %s should be %s", "Description", i, elemt.Description, ref.Elements[i].Description)
		}
		if elemt.Minute != ref.Elements[i].Minute {
			t.Fatalf("Unexpected value for %s at index %d: %d should be %d", "Minute", i, elemt.Minute, ref.Elements[i].Minute)
		}
	}

}

func referenceGameRecorded() *pb.Actions {
	rows := []string{
		"Boston\tJD Davison\t3pts succes\t0",
		"Knicks\tDonte divicenzo\t2pts succes\t0",
		"Knicks\tKevin Mccullar jr\tfree throw try\t1",
		"Knicks\tDonte divicenzo\t3pts succes\t2",
		"Boston\tJD Davison\tfree throw succes\t2",
		"Boston\tJaylen Brown\t3pts try\t4",
		"Boston\tJD Davison\t2pts try\t4",
		"Knicks\tTyler kolek\t2pts try\t6",
		"Knicks\tPacome Dadiet\tfree throw try\t6",
		"Boston\tJaylen Brown\tfree throw succes\t8",
		"Knicks\tDonte divicenzo\tfree throw try\t9",
		"Knicks\tDonte divicenzo\t3pts try\t10",
		"Knicks\tDonte divicenzo\t3pts try\t12",
		"Boston\tSam Hauser\tfree throw succes\t14",
	}

	var ref pb.Actions
	// var action pb.Action
	for _, row := range rows {
		fields := strings.Split(row, "\t")
		if len(fields) == 4 {
			minute := 0
			fmt.Sscanf(fields[3], "%d", &minute)
			action := pb.Action{
				GamePoster:  "testingGame",
				Team:        fields[0],
				PlayerName:  fields[1],
				Description: fields[2],
				Minute:      int32(minute),
			}
			ref.Elements = append(ref.Elements, &action)
		}
	}
	return &ref
}
