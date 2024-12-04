package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"time"

	pb "sync_score/proto"
	sp "sync_score/sport"
	ut "sync_score/utils"

	_ "github.com/mattn/go-sqlite3" // SQLite driver
	"github.com/streadway/amqp"
	"google.golang.org/grpc"

	"google.golang.org/grpc/credentials/insecure"
)

type DBWrapper struct {
	clientDB         *sql.DB
	cachedTableNames map[string]bool
	cachePlayerID    map[string]int
}

var (
	rabbitQueuePort = flag.Int("rabbitQueuePort", 5672, "The rabbitMQ port")
)

func main() {
	
	db := DBWrapper{
		clientDB:         getSQLDB(),
		cachedTableNames: make(map[string]bool),
		cachePlayerID:    make(map[string]int),
	}
	db.initDB()
	// Set up connection using DialContext and UseClient
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
		
		db.sendToTables(action)
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

func (db *DBWrapper) initDB() {
	tableName := "playerStatistic"
	query := fmt.Sprintf("SELECT name FROM sqlite_master WHERE type='table' AND name='%s';", tableName)
	var name string
	err := db.clientDB.QueryRow(query).Scan(&name)
	if err == nil {   // The table exist
		ut.Debugf("Table %s exists.\n", tableName)
		// We initialize the cache with values present in the tables
		mapping := db.queryPlayerIdMap()
		db.cachePlayerID = mapping
	} else {	// The table does not exist
		if err == sql.ErrNoRows {
			ut.Infof("Table %s does not exist. So it is created.\n", tableName)
			query := `CREATE TABLE IF NOT EXISTS playerStatistic (
				id INTEGER PRIMARY KEY,
				playerName STRING,
				twoPointTry INTEGER,
				twoPointSuccess INTEGER,
				threePointTry INTEGER,
				threePointSuccess INTEGER
			);`
			_, err := db.clientDB.Exec(query)
			if err != nil {
				fmt.Println(err)
				ut.Fatal(err)
			}
			// Initialize the cache of player's ID with empty mao
			db.cachePlayerID = make(map[string]int)
		} else {
			fmt.Println(err)
			ut.Fatal(err)
		}
	}
}

func (db *DBWrapper) queryPlayerIdMap() map[string]int {
	mapping := make(map[string]int)
	query := `SELECT id, playerName FROM playerStatistic;`
	rows, err := db.clientDB.Query(query)
	if err != nil {
		ut.Debug(err)
		ut.Fatal(err)
	}
	defer rows.Close()
	var id int
	var playerName string
	for rows.Next() {
		if err := rows.Scan(&id, &playerName); err != nil {
			fmt.Println(err)
			ut.Fatal(err)
		}
		mapping[playerName] = id
	}
	return mapping
}

func (db DBWrapper) addPlayerStat(action sp.Action) {
	fmt.Println(action.Description)
	id, ok := db.cachePlayerID[action.PlayerName]
	if !ok {
		query := `INSERT INTO playerStatistic (playerName, twoPointTry, twoPointSuccess, threePointTry, threePointSuccess)
				  VALUES (?, ?, ?, ?, ?)`
		result, err := db.clientDB.Exec(query, action.PlayerName, 0, 0, 0, 0)
		if err != nil {
			fmt.Println(err)
			panic(err)
		}

		// Get the last inserted row ID
		lastID, err := result.LastInsertId()
		if err != nil {
			panic(err)
		}
		db.cachePlayerID[action.PlayerName] = int(lastID)
		id = int(lastID)
	}

	// query to update from the sp.Action
	switch action.Description {
	case "2pts try":
		updateSQL := `UPDATE playerStatistic SET twoPointTry = twoPointTry + ? WHERE id = ?`
		_, err := db.clientDB.Exec(updateSQL, 1, id)
		if err != nil {
			fmt.Println(err)
			ut.Fatal(err)
		}
	case "2pts succes":

		updateSQL := `UPDATE playerStatistic SET twoPointTry = twoPointTry + ?, twoPointSuccess = twoPointSuccess + ? WHERE id = ?`
		_, err := db.clientDB.Exec(updateSQL, 1, 1, id)
		if err != nil {
			fmt.Println(err)
			ut.Fatal(err)
		}
	case "3pts try":
		updateSQL := `UPDATE playerStatistic SET threePointTry = threePointTry + ? WHERE id = ?`
		_, err := db.clientDB.Exec(updateSQL, 1, id)
		if err != nil {
			fmt.Println(err)
			ut.Fatal(err)
		}
	case "3pts succes":
		updateSQL := `UPDATE playerStatistic SET threePointTry = threePointTry + ?, threePointSuccess = threePointSuccess + ? WHERE id = ?`
		_, err := db.clientDB.Exec(updateSQL, 1, 1, id)
		if err != nil {
			fmt.Println(err)
			ut.Fatal(err)
		}
	default:
		fmt.Printf("Ignored for now %s \n", action.Description)
	}

}

func (db *DBWrapper) sendToTables(action sp.Action) {
	// Log the received event
	// create the table for a particular game: teamA_teamB, TODO: could add a date after on the table name
	// add the element
	// TODO: use placeholder ?
	db.addEntryToPerGameTable(action)

	// Send to tables for players statistic.
	db.addPlayerStat(action)

}

func (db *DBWrapper) addEntryToPerGameTable(action sp.Action) {
	ut.Debugf("Received event: Game=%s, Team=%s, Player=%s, Description=%s, Time=%d",
		action.GamePoster, action.Team, action.PlayerName, action.Description, action.Minute)

	var query string
	if !db.cachedTableNames[action.GamePoster] {

		query = fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
				team STRING,
				playerName STRING,
				description STRING,
				minute INTEGER
			);`, action.GamePoster)
		_, err := db.clientDB.Exec(query)
		if err != nil {
			ut.Info(err)
			ut.Fatal(err)
		}
		db.cachedTableNames[action.GamePoster] = true
	}

	query = fmt.Sprintf(`INSERT INTO %s (team, playerName, description, minute) 
		VALUES ('%s', '%s', '%s', %d);`, action.GamePoster, action.Team, action.PlayerName, action.Description, action.Minute)

	_, err := db.clientDB.Exec(query)
	if err != nil {
		ut.Fatal(err)
	}
}

func getSQLDB() *sql.DB {
	// Open a connection pool to SQLite database
	db, err := sql.Open("sqlite3", "./games.db")
	if err != nil {
		fmt.Println(err)
		ut.Fatal(err)
	}
	return db
}


// func main() {
	
// 	// Set up connection using DialContext and UseClient
// 	conn, err := grpc.NewClient("localhost:50051",
// 		grpc.WithTransportCredentials(insecure.NewCredentials()),
// 	)
// 	if err != nil {
// 		ut.Fatalf("Failed to connect: %v", err)
// 	}
// 	defer conn.Close()

// 	clientDB := pb.NewGameCenterClient(conn)

// 	connection, err := amqp.Dial(fmt.Sprintf("amqp://guest:guest@localhost:%d/", *rabbitQueuePort))
// 	if err != nil {
// 		panic(err)
// 	}
// 	defer connection.Close()

// 	fmt.Println("Successfully connected to RabbitMQ instance")

// 	// opening a channel over the connection established to interact with RabbitMQ
// 	channel, err := connection.Channel()
// 	if err != nil {
// 		panic(err)
// 	}
// 	defer channel.Close()
// 	// Create a queue if it does not exist.
// 	_, err = channel.QueueDeclare(
// 		"LiveGame", // name
// 		false,      // durable
// 		false,      // auto delete
// 		false,      // exclusive
// 		false,      // no wait
// 		nil,        // args
// 	)
// 	if err != nil {
// 		panic(err)
// 	}
// 	//

// 	// declaring a consumer with its properties over channel opened
// 	msgs, err := channel.Consume(
// 		"LiveGame", // queue
// 		"",         // consumer
// 		true,       // auto ack
// 		false,      // exclusive
// 		false,      // no local
// 		false,      // no wait
// 		nil,        //args
// 	)
// 	if err != nil {
// 		panic(err)
// 	}

// 	// print consumed messages from queue
// 	forever := make(chan bool)
// 	go func() {
// 		var action sp.Action
// 		for msg := range msgs {
// 			// ut.Printf("msg.Body: %s", msg.Body)
// 			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
// 			defer cancel()
// 			if err := json.Unmarshal(msg.Body, &action); err != nil {
// 				panic(err)
// 			}
// 			_, err := sendToDb(ctx, clientDB, action)
// 			if err != nil {
// 				ut.Fatalf("Error sending event: %v", err)
// 			}
// 			ut.Infof("Game: %s \n \t Team: %s \n \t name of the player: %s \n \t description: %s \n \t time in minute: %d \n",
// 				action.GamePoster, action.Team, action.PlayerName, action.Description, action.Minute)

// 		}
// 	}()

// 	fmt.Println("Waiting for messages...")
// 	<-forever
// }

// func sendToDb(ctx context.Context,  client pb.GameCenterClient, act sp.Action) (*pb.ActionReply, error) {

// 	// Create a game event
// 	event := &pb.Action{
// 		GamePoster:  act.GamePoster,
// 		Team:        act.Team,
// 		PlayerName:  act.PlayerName,
// 		Description: act.Description,
// 		Minute:      act.Minute,
// 	}
	
// 	// Send event to server
// 	response, err := client.SendGameAction(ctx, event)
// 	if err != nil {
// 		ut.Fatalf("Error sending event: %v", err)
// 	}
// 	return response, nil
// }