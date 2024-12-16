package main

import (
	sp "statistic-syncer/sport"
	"sync"
	"testing"
	"time"
)

func TestUpdateCache(t *testing.T) {
	tests := map[string]struct {
		initialCache map[string]sp.ScoreRecord   
		action      sp.Action
		wantGames   map[string]sp.ScoreRecord
	}{
		"New game scores 2pts team A": {
			initialCache: map[string]sp.ScoreRecord{},  // empty cache
			action: sp.Action{
				GamePoster:  "Boston_Knicks",
				Team:        "Boston",
				Description: "2pts succes",
			},
			wantGames: map[string]sp.ScoreRecord{
				"Boston_Knicks": {
					GameName: "Boston_Knicks",
					TeamA:    "Boston",
					TeamB:    "Knicks",
					ScoreA:   2,
					ScoreB:   0,
				},
			},
		},
		"New game scores 2pts Team B": {
			initialCache: map[string]sp.ScoreRecord{},  // empty cache
			action: sp.Action{
				GamePoster:  "Boston_Knicks",
				Team:        "Knicks",
				Description: "2pts succes",
			},
			wantGames: map[string]sp.ScoreRecord{
				"Boston_Knicks": {
					GameName: "Boston_Knicks",
					TeamA:    "Boston",
					TeamB:    "Knicks",
					ScoreA:   0,
					ScoreB:   2,
				},
			},
		},
		"Free throw try should not update score Team A": {
			initialCache: map[string]sp.ScoreRecord{
				"Boston_Knicks": {
					GameName: "Boston_Knicks",
					TeamA:    "Boston",
					TeamB:    "Knicks",
					ScoreA:   2,
					ScoreB:   0,
				},
			},
			action: sp.Action{
				GamePoster:  "Boston_Knicks",
				Team:        "Boston",
				Description: "free throw try",
			},
			wantGames: map[string]sp.ScoreRecord{
				"Boston_Knicks": {
					GameName: "Boston_Knicks",
					TeamA:    "Boston",
					TeamB:    "Knicks",
					ScoreA:   2,
					ScoreB:   0,
				},
			},
		},
		"Free throw try should not update score Team B": {
			initialCache: map[string]sp.ScoreRecord{
				"Boston_Knicks": {
					GameName: "Boston_Knicks",
					TeamA:    "Boston",
					TeamB:    "Knicks",
					ScoreA:   2,
					ScoreB:   0,
				},
			},
			action: sp.Action{
				GamePoster:  "Boston_Knicks",
				Team:        "Knicks",
				Description: "free throw try",
			},
			wantGames: map[string]sp.ScoreRecord{
				"Boston_Knicks": {
					GameName: "Boston_Knicks",
					TeamA:    "Boston",
					TeamB:    "Knicks",
					ScoreA:   2,
					ScoreB:   0,
				},
			},
		},
		"2pts try should not update score Team A": {
			initialCache: map[string]sp.ScoreRecord{
				"Boston_Knicks": {
					GameName: "Boston_Knicks",
					TeamA:    "Boston",
					TeamB:    "Knicks",
					ScoreA:   2,
					ScoreB:   0,
				},
			},
			action: sp.Action{
				GamePoster:  "Boston_Knicks",
				Team:        "Boston",
				Description: "2pts try",
			},
			wantGames: map[string]sp.ScoreRecord{
				"Boston_Knicks": {
					GameName: "Boston_Knicks",
					TeamA:    "Boston",
					TeamB:    "Knicks",
					ScoreA:   2,
					ScoreB:   0,
				},
			},
		},
		"2pts try should not update score Team B": {
			initialCache: map[string]sp.ScoreRecord{
				"Boston_Knicks": {
					GameName: "Boston_Knicks",
					TeamA:    "Boston",
					TeamB:    "Knicks",
					ScoreA:   2,
					ScoreB:   0,
				},
			},
			action: sp.Action{
				GamePoster:  "Boston_Knicks",
				Team:        "Knicks",
				Description: "2pts try",
			},
			wantGames: map[string]sp.ScoreRecord{
				"Boston_Knicks": {
					GameName: "Boston_Knicks",
					TeamA:    "Boston",
					TeamB:    "Knicks",
					ScoreA:   2,
					ScoreB:   0,
				},
			},
		},
		"3pts try should not update score Team A": {
			initialCache: map[string]sp.ScoreRecord{
				"Boston_Knicks": {
					GameName: "Boston_Knicks",
					TeamA:    "Boston",
					TeamB:    "Knicks",
					ScoreA:   2,
					ScoreB:   0,
				},
			},
			action: sp.Action{
				GamePoster:  "Boston_Knicks",
				Team:        "Boston",
				Description: "3pts try",
			},
			wantGames: map[string]sp.ScoreRecord{
				"Boston_Knicks": {
					GameName: "Boston_Knicks",
					TeamA:    "Boston",
					TeamB:    "Knicks",
					ScoreA:   2,
					ScoreB:   0,
				},
			},
		},
		"3pts try should not update score Team B": {
			initialCache: map[string]sp.ScoreRecord{
				"Boston_Knicks": {
					GameName: "Boston_Knicks",
					TeamA:    "Boston",
					TeamB:    "Knicks",
					ScoreA:   2,
					ScoreB:   0,
				},
			},
			action: sp.Action{
				GamePoster:  "Boston_Knicks",
				Team:        "Knicks",
				Description: "3pts try",
			},
			wantGames: map[string]sp.ScoreRecord{
				"Boston_Knicks": {
					GameName: "Boston_Knicks",
					TeamA:    "Boston",
					TeamB:    "Knicks",
					ScoreA:   2,
					ScoreB:   0,
				},
			},
		},
		"Free throw success adds 1 point team A": {
			initialCache: map[string]sp.ScoreRecord{
				"Boston_Knicks": {
					GameName: "Boston_Knicks",
					TeamA:    "Boston",
					TeamB:    "Knicks",
					ScoreA:   2,
					ScoreB:   0,
				},
			},
			action: sp.Action{
				GamePoster:  "Boston_Knicks",
				Team:        "Boston",
				Description: "free throw succes",
			},
			wantGames: map[string]sp.ScoreRecord{
				"Boston_Knicks": {
					GameName: "Boston_Knicks",
					TeamA:    "Boston",
					TeamB:    "Knicks",
					ScoreA:   3,
					ScoreB:   0,
				},
			},
		},
		"Free throw success adds 1 point Team B": {
			initialCache: map[string]sp.ScoreRecord{
				"Boston_Knicks": {
					GameName: "Boston_Knicks",
					TeamA:    "Boston",
					TeamB:    "Knicks",
					ScoreA:   2,
					ScoreB:   0,
				},
			},
			action: sp.Action{
				GamePoster:  "Boston_Knicks",
				Team:        "Knicks",
				Description: "free throw succes",
			},
			wantGames: map[string]sp.ScoreRecord{
				"Boston_Knicks": {
					GameName: "Boston_Knicks",
					TeamA:    "Boston",
					TeamB:    "Knicks",
					ScoreA:   2,
					ScoreB:   1,
				},
			},
		},
		"2pts success adds 2 points Team A": {
			initialCache: map[string]sp.ScoreRecord{
				"Boston_Knicks": {
					GameName: "Boston_Knicks",
					TeamA:    "Boston",
					TeamB:    "Knicks",
					ScoreA:   2,
					ScoreB:   0,
				},
			},
			action: sp.Action{
				GamePoster:  "Boston_Knicks",
				Team:        "Boston",
				Description: "2pts succes",
			},
			wantGames: map[string]sp.ScoreRecord{
				"Boston_Knicks": {
					GameName: "Boston_Knicks",
					TeamA:    "Boston",
					TeamB:    "Knicks",
					ScoreA:   4,
					ScoreB:   0,
				},
			},
		},
		"2pts success adds 2 points Team B": {
			initialCache: map[string]sp.ScoreRecord{
				"Boston_Knicks": {
					GameName: "Boston_Knicks",
					TeamA:    "Boston",
					TeamB:    "Knicks",
					ScoreA:   2,
					ScoreB:   0,
				},
			},
			action: sp.Action{
				GamePoster:  "Boston_Knicks",
				Team:        "Knicks",
				Description: "2pts succes",
			},
			wantGames: map[string]sp.ScoreRecord{
				"Boston_Knicks": {
					GameName: "Boston_Knicks",
					TeamA:    "Boston",
					TeamB:    "Knicks",
					ScoreA:   2,
					ScoreB:   2,
				},
			},
		},
		"3pts success adds 3 points Team A": {
			initialCache: map[string]sp.ScoreRecord{
				"Boston_Knicks": {
					GameName: "Boston_Knicks",
					TeamA:    "Boston",
					TeamB:    "Knicks",
					ScoreA:   2,
					ScoreB:   0,
				},
			},
			action: sp.Action{
				GamePoster:  "Boston_Knicks",
				Team:        "Boston",
				Description: "3pts succes",
			},
			wantGames: map[string]sp.ScoreRecord{
				"Boston_Knicks": {
					GameName: "Boston_Knicks",
					TeamA:    "Boston",
					TeamB:    "Knicks",
					ScoreA:   5,
					ScoreB:   0,
				},
			},
		},
		"3pts success adds 3 points Team B": {
			initialCache: map[string]sp.ScoreRecord{
				"Boston_Knicks": {
					GameName: "Boston_Knicks",
					TeamA:    "Boston",
					TeamB:    "Knicks",
					ScoreA:   2,
					ScoreB:   0,
				},
			},
			action: sp.Action{
				GamePoster:  "Boston_Knicks",
				Team:        "Knicks",
				Description: "3pts succes",
			},
			wantGames: map[string]sp.ScoreRecord{
				"Boston_Knicks": {
					GameName: "Boston_Knicks",
					TeamA:    "Boston",
					TeamB:    "Knicks",
					ScoreA:   2,
					ScoreB:   3,
				},
			},
		},
		"Foul should not update score": {
			initialCache: map[string]sp.ScoreRecord{
				"Boston_Knicks": {
					GameName: "Boston_Knicks",
					TeamA:    "Boston",
					TeamB:    "Knicks",
					ScoreA:   2,
					ScoreB:   0,
				},
			},
			action: sp.Action{
				GamePoster:  "Boston_Knicks",
				Team:        "Boston",
				Description: "foul",
			},
			wantGames: map[string]sp.ScoreRecord{
				"Boston_Knicks": {
					GameName: "Boston_Knicks",
					TeamA:    "Boston",
					TeamB:    "Knicks",
					ScoreA:   2,
					ScoreB:   0,
				},
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			cache := NewCacheGameRecorded()
			cache.games = tc.initialCache

			cache.updateCache(tc.action)

			if len(cache.games) != len(tc.wantGames) {
				t.Errorf("got %d games, want %d games", len(cache.games), len(tc.wantGames))
			}

			for gameName, wantGame := range tc.wantGames {
				gotGame, exists := cache.games[gameName]
				if !exists {
					t.Fatalf("game %s not found in cache", gameName)
				}
				if gotGame != wantGame {
					t.Errorf("game %s = %+v, want %+v", gameName, gotGame, wantGame)
				}
			}
		})
	}
}

func TestGetScore(t *testing.T) {
	tests := map[string]struct {
		initialCache map[string]sp.ScoreRecord
		gamePoster  string
		want        sp.ScoreRecord
	}{
		"Get score for existing game": {
			initialCache: map[string]sp.ScoreRecord{
				"Boston_Knicks": {
					GameName: "Boston_Knicks",
					TeamA:    "Boston",
					TeamB:    "Knicks",
					ScoreA:   10,
					ScoreB:   5,
				},
			},
			gamePoster: "Boston_Knicks",
			want: sp.ScoreRecord{
				GameName: "Boston_Knicks",
				TeamA:    "Boston",
				TeamB:    "Knicks",
				ScoreA:   10,
				ScoreB:   5,
			},
		},
		"Get score for non-existing game": {
			initialCache: map[string]sp.ScoreRecord{},
			gamePoster:  "NonExisting_Game",
			want:        sp.ScoreRecord{},
		},
		"Get score with multiple games in cache": {
			initialCache: map[string]sp.ScoreRecord{
				"Boston_Knicks": {
					GameName: "Boston_Knicks",
					TeamA:    "Boston",
					TeamB:    "Knicks",
					ScoreA:   10,
					ScoreB:   5,
				},
				"Lakers_Bulls": {
					GameName: "Lakers_Bulls",
					TeamA:    "Lakers",
					TeamB:    "Bulls",
					ScoreA:   15,
					ScoreB:   12,
				},
			},
			gamePoster: "Lakers_Bulls",
			want: sp.ScoreRecord{
				GameName: "Lakers_Bulls",
				TeamA:    "Lakers",
				TeamB:    "Bulls",
				ScoreA:   15,
				ScoreB:   12,
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			cache := NewCacheGameRecorded()
			cache.games = tc.initialCache

			// Test concurrent access for each test case
			const numGoroutines = 10
			var wg sync.WaitGroup
			wg.Add(numGoroutines)

			for i := 0; i < numGoroutines; i++ {
				go func() {
					defer wg.Done()
					got := cache.getScore(tc.gamePoster)
					if got != tc.want {
						t.Errorf("getScore() = %v, want %v", got, tc.want)
					}
				}()
			}

			wg.Wait()
		})
	}
}

func Test_updateCache_concurent(t *testing.T) {
	cache := NewCacheGameRecorded()
    
    // Initialize cache with a game record
    initialGame := sp.ScoreRecord{
        GameName: "Boston_Knicks",
        TeamA:    "Boston",
        TeamB:    "Knicks",
        ScoreA:   0,
        ScoreB:   0,
    }
    cache.games = map[string]sp.ScoreRecord{
        "Boston_Knicks": initialGame,
    }

    // Define test actions
    actions := []sp.Action{
        {
            GamePoster:  "Boston_Knicks",
            Team:        "Boston",
            Description: "2pts succes",
        },
        {
            GamePoster:  "Boston_Knicks",
            Team:        "Knicks",
            Description: "3pts succes",
        },
        {
            GamePoster:  "Boston_Knicks",
            Team:        "Boston",
            Description: "free throw succes",
        },
    }

    // Run concurrent updates
    const numGoroutines = 10
    var wg sync.WaitGroup
    wg.Add(numGoroutines)

    for i := 0; i < numGoroutines; i++ {
        go func(idx int) {
            defer wg.Done()
            // Each goroutine will execute all actions
            for _, action := range actions {
                cache.updateCache(action)
            }
        }(i)
    }

    wg.Wait()

    // Verify final state
    finalScore := cache.getScore("Boston_Knicks")
    
    // Each goroutine adds: 2 + 3 + 1 = 6 points for each team
    // Total points per team = 6 * numGoroutines
    expectedPointsPerTeam := int32(3 * numGoroutines)
    
    if finalScore.ScoreA != expectedPointsPerTeam {
        t.Errorf("Team A final score = %d, want %d", finalScore.ScoreA, expectedPointsPerTeam)
    }
    if finalScore.ScoreB != expectedPointsPerTeam {
        t.Errorf("Team B final score = %d, want %d", finalScore.ScoreB, expectedPointsPerTeam)
    }

}

func TestClearCacheIfExpired(t *testing.T) {
	// Use a short TTL for testing
	ttl := 100 * time.Millisecond

	tests := map[string]struct {
		setup    func(*CacheGameRecorded)
		validate func(*testing.T, *CacheGameRecorded)
	}{
		"Expired entries should be removed": {
			setup: func(c *CacheGameRecorded) {
				expiredTime := time.Now().Add(-2 * ttl)
				c.games = map[string]sp.ScoreRecord{
					"Game1": {
						GameName: "Game1",
						LastRead: expiredTime,
					},
					"Game2": {
						GameName: "Game2",
						LastRead: expiredTime,
					},
				}
			},
			validate: func(t *testing.T, c *CacheGameRecorded) {
				if len(c.games) != 0 {
					t.Errorf("Expected all expired entries to be removed, got %d entries", len(c.games))
				}
			},
		},
		"Valid entries should be kept": {
			setup: func(c *CacheGameRecorded) {
				validTime := time.Now()
				c.games = map[string]sp.ScoreRecord{
					"Game1": {
						GameName: "Game1",
						LastRead: validTime,
					},
					"Game2": {
						GameName: "Game2",
						LastRead: validTime,
					},
				}
			},
			validate: func(t *testing.T, c *CacheGameRecorded) {
				if len(c.games) != 2 {
					t.Errorf("Expected valid entries to be kept, got %d entries", len(c.games))
				}
			},
		},
		"Mixed expired and valid entries": {
			setup: func(c *CacheGameRecorded) {
				validTime := time.Now()
				expiredTime := time.Now().Add(-2 * ttl)
				c.games = map[string]sp.ScoreRecord{
					"ExpiredGame": {
						GameName: "ExpiredGame",
						LastRead: expiredTime,
					},
					"ValidGame": {
						GameName: "ValidGame",
						LastRead: validTime,
					},
				}
			},
			validate: func(t *testing.T, c *CacheGameRecorded) {
				if len(c.games) != 1 {
					t.Errorf("Expected only one valid entry to remain, got %d entries", len(c.games))
				}
				if _, exists := c.games["ValidGame"]; !exists {
					t.Error("Valid game should still exist in cache")
				}
				if _, exists := c.games["ExpiredGame"]; exists {
					t.Error("Expired game should have been removed from cache")
				}
			},
		},
		"Future timestamps should be reset": {
			setup: func(c *CacheGameRecorded) {
				futureTime := time.Now().Add(24 * time.Hour)
				c.games = map[string]sp.ScoreRecord{
					"FutureGame": {
						GameName: "FutureGame",
						LastRead: futureTime,
						ScoreA:   10,
						ScoreB:   20,
					},
				}
			},
			validate: func(t *testing.T, c *CacheGameRecorded) {
				if game, exists := c.games["FutureGame"]; exists {
					if !game.LastRead.Before(time.Now().Add(time.Second)) {
						t.Error("Future timestamp should have been reset to current time")
					}
					// Verify other data remains unchanged
					if game.ScoreA != 10 || game.ScoreB != 20 {
						t.Error("Game scores should remain unchanged after timestamp reset")
					}
				} else {
					t.Error("Future game should still exist in cache")
				}
			},
		},
		"Empty cache should not cause errors": {
			setup: func(c *CacheGameRecorded) {
				c.games = map[string]sp.ScoreRecord{}
			},
			validate: func(t *testing.T, c *CacheGameRecorded) {
				if len(c.games) != 0 {
					t.Error("Empty cache should remain empty")
				}
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			cache := NewCacheGameRecorded(ttl)
			tc.setup(cache)
			
			// Execute clearCacheIfExpired
			cache.clearCacheIfExpired()
			
			// Validate results
			tc.validate(t, cache)
		})
	}
}
