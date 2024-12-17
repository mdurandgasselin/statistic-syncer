package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand/v2"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
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

type Game struct {
	ID string `json:"id"`
	Data GameData `json:"data"`
	Started bool `json:"started"`
	Done chan bool `json:"done"`
}

type RouterRegistry struct {
	router *mux.Router
	mu     sync.RWMutex
}

func NewRouterRegistry() *RouterRegistry {
	return &RouterRegistry{
		router: mux.NewRouter(),
	}
}

func (rr *RouterRegistry) Register(pattern string, handler http.HandlerFunc) {
	rr.mu.Lock()
	defer rr.mu.Unlock()
	rr.router.HandleFunc(pattern, handler)
}

func (rr *RouterRegistry) RegisterWithMethod(pattern, method string, handler http.HandlerFunc) {
	rr.mu.Lock()
	defer rr.mu.Unlock()
	rr.router.HandleFunc(pattern, handler).Methods(method)
}

var (
	games = make(map[string]*Game)
	gamesMux sync.RWMutex
)

// var game GameData
// var gameStarted bool


// Update startGame to use Game instead of global game variable
func startGame(g *Game) {
	fmt.Printf("Starting game %s\n", g.ID)
	for {
		time.Sleep(1 * time.Second)
		select {
		case <-g.Done:
			log.Printf("Game %s done\n", g.ID)
			return
		default:
			if rand.IntN(100) > 50 {
				g.Data.TeamAScore++
				g.Data.Actions = append(g.Data.Actions,
					SportAction{
						Team:                "TeamA",
						PlayerName:          "Player 1",
						DescriptionOfAction: "3 points try",
						IsSuccess:           true},
                    )
			} else {
				g.Data.TeamBScore++
				g.Data.Actions = append(g.Data.Actions,
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


func endGameAfterNSeconds(n int, game *Game) {
	go func() {
		time.Sleep(time.Duration(n) * time.Second)
		game.Done <- true
	}()
}

func main() {
	registry := NewRouterRegistry()

	

	// Initialize games dynamically
	gamesMux.Lock()
	for i := 1; i <= 2; i++ { // You can change this number
		gameID := fmt.Sprintf("game%d", i)
		games[gameID] = initGame(gameID)
	}
	gamesMux.Unlock()

	// Add new game after 15 seconds
    go func() {
        time.Sleep(15 * time.Second)
        gamesMux.Lock()
        games["game3"] = initGame("game3")
        gamesMux.Unlock()
        log.Println("New game 'game3' has been added!")
    }()

	// Root handler
	registry.RegisterWithMethod("/", "GET", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "landing.html")
	})

	// Add games list endpoint
	registry.RegisterWithMethod("/games", "GET", func(w http.ResponseWriter, r *http.Request) {
		gamesMux.RLock()
		gamesList := make([]string, 0, len(games))
		for id := range games {
			gamesList = append(gamesList, id)
		}
		gamesMux.RUnlock()

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(gamesList)
	})

	// Generic game view handler - Add GET method explicitly
	registry.RegisterWithMethod("/game/{id}", "GET", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		gameID := vars["id"]
		
		gamesMux.RLock()
		_, exists := games[gameID]
		gamesMux.RUnlock()
		
		if !exists {
			http.NotFound(w, r)
			return
		}
		http.ServeFile(w, r, "index.html")
	})

	// Generic stream handler - Add GET method explicitly
	registry.RegisterWithMethod("/stream/{id}", "GET", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		gameID := vars["id"]
		
		gamesMux.RLock()
		game, exists := games[gameID]
		gamesMux.RUnlock()
		
		if !exists {
			http.NotFound(w, r)
			return
		}

		// Set headers for SSE
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		// Start game if not started
		if !game.Started {
			game.Started = true
			go startGame(game)
			endGameAfterNSeconds(15, game)
		}

		// Stream data
		for {
			data := GameData{
				CurrentTime: time.Now(),
				GameName:    game.Data.GameName,
				TeamAScore:  game.Data.TeamAScore,
				TeamBScore:  game.Data.TeamBScore,
				Actions:     game.Data.Actions,
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
	server := &http.Server{
		Addr:    ":8080",
		Handler: registry.router,
	}

	// Start the server
	log.Println("Server is running at http://localhost:8080")
	log.Fatal(server.ListenAndServe())
}

func initGame(id string) *Game {
	return &Game{
		ID: id,
		Data: GameData{
			CurrentTime: time.Now(),
			GameName:    fmt.Sprintf("Exciting Match %s", id),
			TeamAScore:  0,
			TeamBScore:  0,
			Actions:     make([]SportAction, 0),
		},
		Done: make(chan bool),
	}
}
