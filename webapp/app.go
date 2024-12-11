package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand/v2"
	"net/http"
	"time"
)

type SportAction struct {
	Team                string `json:"team"`
	PlayerName          string `json:"playerName"`
	DescriptionOfAction string `json:"descriptionOfAction"`
	IsSuccess           bool   `json:"isSuccess"`
}

type GameData struct {
	CurrentTime time.Time     `json:"currentTime"`
	GameName    string        `json:"gameName"`
	TeamAScore  int           `json:"teamAScore"`
	TeamBScore  int           `json:"teamBScore"`
	Actions     []SportAction `json:"actions"`
}

var game GameData

func startGame(done <-chan bool) {
	for {
		time.Sleep(1 * time.Second)
		select {
		case <-done: // iF THE GAME must END SOONER
			return
		default:
			if rand.IntN(100) > 50 {
				game.TeamAScore++
				game.Actions = append(game.Actions,
					SportAction{
						Team:                "TeamA",
						PlayerName:          "Player 1",
						DescriptionOfAction: "3 points try",
						IsSuccess:           true},
                    )

			} else {
				game.TeamBScore++
                game.Actions = append(game.Actions,
					SportAction{
						Team:                "TeamA",
						PlayerName:          "Player 2",
						DescriptionOfAction: "2 points try",
						IsSuccess:           false},
                    )
			}
		}
	}
}

func startNsecond(n int) chan bool {
	done := make(chan bool)
	go func() {
		time.Sleep(time.Duration(n) * time.Second)
		done <- true
	}()
	return done
}

func main() {
	initGame()
	done := startNsecond(20)
	go startGame(done)
	// Serve the HTML file
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})

	// Handle the /stream endpoint
	http.HandleFunc("/stream", func(w http.ResponseWriter, r *http.Request) {
		// Set headers for SSE
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		// Stream data to the client
		for {
			data := GameData{
				CurrentTime: time.Now(),
				GameName:    game.GameName,
				TeamAScore:  game.TeamAScore,
				TeamBScore:  game.TeamBScore,
				Actions: game.Actions,
			}
			jsonData, err := json.Marshal(data)
			if err != nil {
				log.Println("Error marshaling JSON:", err)
				continue
			}
			fmt.Fprintf(w, "data: %s\n\n", jsonData)
			w.(http.Flusher).Flush()
			time.Sleep(1 * time.Second)

		}
	})

	// Start the server
	log.Println("Server is running at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func initGame() {
	game = GameData{
		CurrentTime: time.Now(),
		GameName:    "Exciting Match",
		TeamAScore:  0,
		TeamBScore:  0,
		Actions:     make([]SportAction, 0),
	}
}
