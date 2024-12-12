package db

import (
	"database/sql"
	"fmt"

	ut "sync_score/utils"
	sp "sync_score/sport"

	_ "github.com/mattn/go-sqlite3" // SQLite driver
)

type DBWrapper struct {
	clientDB         *sql.DB
	cachedTableNames map[string]bool
	cachePlayerID    map[string]int
}

func NewDBWrapper(dbName string) DBWrapper{
	db := DBWrapper{
		clientDB:         getSQLDB(dbName),
		cachedTableNames: make(map[string]bool),
		cachePlayerID:    make(map[string]int),
	}
	db.initDB()
	return db
}

func (db *DBWrapper) initDB() {
	tableName := "playerStatistic"
	query := fmt.Sprintf("SELECT name FROM sqlite_master WHERE type='table' AND name='%s';", tableName)
	var name string
	err := db.clientDB.QueryRow(query).Scan(&name)
	if err == nil { // The table exist
		ut.Debugf("Table %s exists.\n", tableName)
		// We initialize the cache with values present in the tables
		mapping := db.queryPlayerIdMap()
		db.cachePlayerID = mapping
	} else { // The table does not exist
		if err == sql.ErrNoRows {
			ut.Infof("Table %s does not exist. So it is created.\n", tableName)
			query := `CREATE TABLE IF NOT EXISTS playerStatistic (
				id INTEGER PRIMARY KEY,
				playerName STRING,
				twoPointTry INTEGER,
				twoPointSuccess INTEGER,
				threePointTry INTEGER,
				threePointSuccess INTEGER,
				freeThrowTry INTEGER,
				freeThrowSuccess INTEGER,
				foul INTEGER
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

func (db *DBWrapper) QueryGameHistoric(gamePoster string) (sp.Actions, error) {
	query := fmt.Sprintf(`SELECT * FROM %s;`, gamePoster)

	rows, err := db.clientDB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	// Extract Values
	var actions sp.Actions
	var team, playerName, description string
	var minute int32
	for rows.Next() {
		if err := rows.Scan(&team, &playerName, &description, &minute); err != nil {
			ut.Fatal(err)
		}
		actions = append(actions, sp.Action{
			GamePoster:  gamePoster,
			Team:        team,
			PlayerName:  playerName,
			Description: description,
			Minute:      minute})
	}
	return actions, nil
}

func (db DBWrapper) addPlayerStat(action sp.Action) {
	fmt.Println(action.Description)
	id, ok := db.cachePlayerID[action.PlayerName]
	if !ok {
		query := `INSERT INTO playerStatistic (playerName, twoPointTry, twoPointSuccess, threePointTry, threePointSuccess, freeThrowTry, freeThrowSuccess, foul)
				  VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
		result, err := db.clientDB.Exec(query, action.PlayerName, 0, 0, 0, 0, 0, 0, 0)
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
		
	case "free throw succes":
		updateSQL := `UPDATE playerStatistic SET freeThrowSuccess = freeThrowSuccess + ?, freeThrowTry = freeThrowTry + ? WHERE id = ?`
		_, err := db.clientDB.Exec(updateSQL, 1, 1, id)
		if err != nil {
			fmt.Println(err)
			ut.Fatal(err)
		}
	case "free throw try":
		updateSQL := `UPDATE playerStatistic SET freeThrowTry = freeThrowTry + ? WHERE id = ?`
		_, err := db.clientDB.Exec(updateSQL, 1, id)
		if err != nil {
			fmt.Println(err)
			ut.Fatal(err)
		}
	case  "foul":
		updateSQL := `UPDATE playerStatistic SET foul = foul + ? WHERE id = ?`
		_, err := db.clientDB.Exec(updateSQL, 1, id)
		if err != nil {
			fmt.Println(err)
			ut.Fatal(err)
		}
	default:
		fmt.Printf("Ignored for now %s \n", action.Description)
	}

}

func (db *DBWrapper) SendToTables(action sp.Action) {
	// send to per game DB	
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

func (db *DBWrapper) queryGameRecord(gameName string) []sp.Action {
	// var query string
	// if !db.cachedTableNames[gameName] {

	// 	query = fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
	// 			team STRING,
	// 			playerName STRING,
	// 			description STRING,
	// 			minute INTEGER
	// 		);`, gameName)
	// 	_, err := db.clientDB.Exec(query)
	// 	if err != nil {
	// 		ut.Info(err)
	// 		ut.Fatal(err)
	// 	}
	// 	db.cachedTableNames[action.GamePoster] = true
	// }

	// query = fmt.Sprintf(`INSERT INTO %s (team, playerName, description, minute) 
	// 	VALUES ('%s', '%s', '%s', %d);`, gameName, action.Team, action.PlayerName, action.Description, action.Minute)

	// _, err := db.clientDB.Exec(query)
	// if err != nil {
	// 	ut.Fatal(err)
	// }
	return make([]sp.Action, 0)
}

func getSQLDB(dbName string) *sql.DB {
	// Open a connection pool to SQLite database
	db, err := sql.Open("sqlite3", dbName) //"./games.db")
	if err != nil {
		fmt.Println(err)
		ut.Fatal(err)
	}
	return db
}
