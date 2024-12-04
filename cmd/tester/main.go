package main

import (
	"database/sql"
	"fmt"
	"log"

	ut "sync_score/utils"

	_ "github.com/mattn/go-sqlite3" // SQLite driver
)



func main() {
	db, err := sql.Open("sqlite3", "./monolithe/games.db")
	if err != nil {
		fmt.Println(err)
		ut.Fatal(err)
	}

	showRowsInTable(db)

	getDbTableNames(db)
	
}

func convertToStringSlice(slice []any) ([]string, error) {
    result := make([]string, 0, len(slice))
    
    for _, v := range slice {
        str, ok := v.(string)
        if !ok {
            return nil, fmt.Errorf("element is not a string: %v", v)
        }
        result = append(result, str)
    }
    
    return result, nil
}

func getUniqueValues(db *sql.DB, colName, table string) []any {
	// Query to get unique values from a column
	query := fmt.Sprintf("SELECT DISTINCT %s FROM %s;", colName, table)

	// Execute the query
	rows, err := db.Query(query)
	if err != nil {
		fmt.Println(err)
		ut.Fatal(err)
	}
	defer rows.Close()

	// Slice to store the unique values
	var uniqueValues []any

	// Iterate over the rows and extract the values
	for rows.Next() {
		var value any
		if err := rows.Scan(&value); err != nil {
			ut.Fatal(err)
		}
		uniqueValues = append(uniqueValues, value)
	}

	// Check for errors from iterating over rows
	if err := rows.Err(); err != nil {
		ut.Fatal(err)
	}
	return uniqueValues	
}


func showRowsInTable(db *sql.DB) {
	query := "SELECT * FROM playerStatistic;"
	rows, err := db.Query(query)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	fmt.Println("rows in table: ")
	for rows.Next() {
		var playerName string
		var id, n2pts, succes2pts, n3pts, success3pts int
		if err := rows.Scan(&id, &playerName, &n2pts, &succes2pts, &n3pts, &success3pts); err != nil {
			fmt.Println(err)
			ut.Fatal(err)
		}

		fmt.Printf("id: %d, playerName: %s, n2pts: %d, succes2pts: %d, n3pts: %d, success3pts: %d \n", id, playerName, n2pts, succes2pts, n3pts, success3pts)
	}
}	

func getDbTableNames(db *sql.DB) ([]string) {
	query := "SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%';"

	rows, err := db.Query(query)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	fmt.Println("Tables in games.db:")
	var names []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			ut.Fatal(err)
		}
		names = append(names, tableName)
		fmt.Println(tableName)
	}
	
	if err := rows.Err(); err != nil {
		ut.Fatal(err)
	}
	return names
}


func myTest() {
    // Create a map to cache the presence of elements
    cache := make(map[string]bool)

    // Add elements to the cache
    cache["apple"] = true
    cache["banana"] = true
    cache["cherry"] = true

    // Check if an element is in the cache
    element := "banana"
    if cache[element] {
        fmt.Println(element, "is in the cache")
    } else {
        fmt.Println(element, "is not in the cache")
    }

    // Check another element
    element = "grape"
    if cache[element] {
        fmt.Println(element, "is in the cache")
    } else {
        fmt.Println(element, "is not in the cache")
    }
}