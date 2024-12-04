package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log"

	sp "sync_score/sport"
	ut "sync_score/utils"

	_ "github.com/mattn/go-sqlite3" // SQLite driver
	"github.com/streadway/amqp"
)

type DBWrapper struct {
	clientDB         *sql.DB
	cachedTableNames map[string]bool
	cachePlayerID    map[string]int
}

type QueueWrapper struct {
	queueChannel *amqp.Channel
	cache        map[string]bool
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
	// create the player statistic table.
	

	connection, err := amqp.Dial(fmt.Sprintf("amqp://guest:guest@localhost:%d/", *rabbitQueuePort))
	if err != nil {
		panic(err)
	}
	defer connection.Close()
	ut.Info("Successfully connected to RabbitMQ instance")

	// opening a channel over the connection established to interact with RabbitMQ
	channel, err := connection.Channel()
	if err != nil {
		panic(err)
	}
	defer channel.Close()

	// Create a queue if it does not exist.
	_, err = channel.QueueDeclare(
		"LiveGame", // name
		false,      // durable
		false,      // auto delete
		false,      // exclusive
		false,      // no wait
		nil,        // args
	)
	if err != nil {
		panic(err)
	}

	// declaring a consumer with its properties over channel opened
	msgs, err := channel.Consume(
		"LiveGame", // queue
		"",         // consumer
		true,       // auto ack
		false,      // exclusive
		false,      // no local
		false,      // no wait
		nil,        //args
	)
	if err != nil {
		panic(err)
	}

	var action sp.Action
	for msg := range msgs {
		if err := json.Unmarshal(msg.Body, &action); err != nil {
			panic(err)
		}
		ut.Infof("Game: %s \n \t Team: %s \n \t name of the player: %s \n \t description: %s \n \t time in minute: %d \n",
			action.GamePoster, action.Team, action.PlayerName, action.Description, action.Minute)
		// send to DB for later update
		db.sendToDb(action)
		
		// Send to queue to keep update on live statistic
		db.compile(action)
	}
}

func (db *DBWrapper) initDB() {
	// Check if 
	tableName := "playerStatistic"
    query := fmt.Sprintf("SELECT name FROM sqlite_master WHERE type='table' AND name='%s';", tableName)
    var name string
    err := db.clientDB.QueryRow(query).Scan(&name)
    if err == nil {
		ut.Infof("Table %s exists.\n", tableName)
		// We initialize the cache with values present in the tables
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
			db.cachePlayerID[playerName] = id
		}
	} else {
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
		return
	}

}

func (db DBWrapper) compile(action sp.Action) {
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

func (q QueueWrapper) sendToQueue(action sp.Action) {
	var err error
	if !q.cache[action.GamePoster] {
		// Create a queue if it does not exist.
		_, err = q.queueChannel.QueueDeclare(
			action.GamePoster, // name
			false,             // durable
			false,             // auto delete
			false,             // exclusive
			false,             // no wait
			nil,               // args
		)
		if err != nil {
			panic(err)
		}
		q.cache[action.GamePoster] = true
	}
	// publishing a message
	body, _ := json.Marshal(action)
	err = q.queueChannel.Publish(
		"",                // exchange
		action.GamePoster, // key
		false,             // mandatory
		false,             // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        body,
		},
	)
	if err != nil {
		log.Fatalf("could not publish: %v", err)
	}
}

func (db *DBWrapper) sendToDb(action sp.Action) {
	// Log the received event
	ut.Infof("Received event: Game=%s, Team=%s, Player=%s, Description=%s, Time=%d",
		action.GamePoster, action.Team, action.PlayerName, action.Description, action.Minute)

	var query string
	if !db.cachedTableNames[action.GamePoster] {
		// create the table for particule game: teamA_teamB
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

	// add the element
	query = fmt.Sprintf(`INSERT INTO %s (team, playerName, description, minute) 
		VALUES ('%s', '%s', '%s', %d);`, action.GamePoster, action.Team, action.PlayerName, action.Description, action.Minute)
	// TODO: use placeholder ?
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
