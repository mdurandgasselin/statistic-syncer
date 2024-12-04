package sport

import (
	"encoding/json"
	"os"

	ut "sync_score/utils"
)

type Actions []Action

type Action struct {
	GamePoster  string `json:"gameposter"`
	Team        string `json:"team"`
	PlayerName  string `json:"playername"`
	Description string `json:"description"`
	Minute      int32  `json:"minute"`
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
