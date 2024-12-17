package sport

import (
	"encoding/json"
	"os"
	"time"

	ut "statistic-syncer/utils"
)

type Actions []Action

type Action struct {
	GamePoster  string `json:"gameposter"`
	Team        string `json:"team"`
	PlayerName  string `json:"playername"`
	Description string `json:"description"`
	Minute      int32  `json:"minute"`
}

type ScoreRecord struct {
	GameName string `json:"gameName"`
	TeamA    string `json:"teamA"`
	TeamB    string `json:"teamB"`
	ScoreA   int32  `json:"scoreA"`
	ScoreB   int32  `json:"scoreB"`
	// Non exported field
	LastRead time.Time `json:"-"`
}

func (s *ScoreRecord) Reset() {
	s.LastRead = time.Now()
}

func ReadGameFile(path string) (Actions, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		ut.Fatal(err)
	}
	var game Actions
	err = json.Unmarshal(content, &game)
	if err != nil {
		ut.Fatal(err)
	}
	return game, nil
}
