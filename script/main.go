package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand/v2"
	"os"
	sp "sync_score/sport"
)

func main() {

	playerBoston := []string{"Jaylen Brown", "JD Davison", "tristan enaruna", "Ron harper", "Sam Hauser"}
	playerKnicks := []string{"Pacome Dadiet", "Tyler kolek", "ariel Hukporti", "Kevin Mccullar jr", "Donte divicenzo"}
	gameBostonKnicks := playGame("Boston", "Knicks", playerBoston, playerKnicks)

	playerSixers := []string{"tyrese Maxey", "Reggie Jackson", "KJ MArtin", "Andre Drummond", "Kyle Lowry"}
	playerRaptor := []string{"JaKobe Walter", "Jonathan Mogbo", "Jamal Shead", "Malik Williams", "Brandon Carlson"}
	gameSixersRaptor := playGame("Sixers", "Raptor", playerSixers, playerRaptor)

	playerbulls := []string{"Marcus Domask", "Coby White", "Lonzo Ball", "Josh Giddey", "Chris Duarte"}
	playerCavalier := []string{"Jaylon Tyson", "Max trus", "Ty Jerome", "Caris Levert", "Evan Mobley"}
	gameBullsCavalier := playGame("Bulls", "cavaliers", playerbulls, playerCavalier)
	allGames := map[string][]BasketEvent{"Boston-Knicks": gameBostonKnicks,
		"Sixers-Raptor":   gameSixersRaptor,
		"Bulls-cavaliers": gameBullsCavalier,
	}

	//...................................
	//Writing struct type to a JSON file
	//...................................
	for gameName, game := range allGames {
		content, err := json.Marshal(game)
		if err != nil {
			fmt.Println(err)
		}
		err = os.WriteFile(fmt.Sprintf("%s.json", gameName), content, 0644)
		if err != nil {
			log.Fatal(err)
		}
	}

	//...................................
	//Reading into struct type from a JSON file
	//...................................
	content, err := os.ReadFile("Sixers-Raptor.json")
	if err != nil {
		log.Fatal(err)
	}
	var game2 []BasketEvent
	err = json.Unmarshal(content, &game2)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Id:%s, Name:%s, Password:%s, LoggedAt:%d",
		 game2[0].Team, game2[0].PlayerName, game2[0].Description, game2[0].Minute)
}

type BasketEvent sp.Action
// struct {
// 	Team       string
// 	PlayerName string
// 	Description     string
// 	Minute       int
// }

func playGame(teamAname string, teamBname string, teamA []string, teamB []string) []BasketEvent {
	maxTime := int32(60)
	currentTime := int32(0)
	actions := []string{"free throw try", "2pts try", "3pts try", "free throw succes", "2pts succes", "3pts succes", "foul"}
	var events []BasketEvent
	for currentTime < maxTime {
		if rand.Float64() > 0.5 {
			events = append(events, BasketEvent{
				GamePoster: teamAname+ "-" + teamBname,
				Team: teamAname,
				PlayerName: teamA[rand.IntN(len(teamA))],
				Description:     actions[rand.IntN(len(actions))],
				Minute: currentTime,
			})
		} else {
			events = append(events, BasketEvent{
				GamePoster: teamAname+ "-" + teamBname,
				Team: teamBname,
				PlayerName: teamB[rand.IntN(len(teamB))],
				Description:     actions[rand.IntN(len(actions))],
				Minute: currentTime,
			})

		}
		currentTime += int32(rand.IntN(3))
	}
	return events
}
