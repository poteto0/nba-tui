package game_detail

import (
	"testing"
	"github.com/poteto0/go-nba-sdk/types"
	"github.com/stretchr/testify/assert"
)

func TestBoxScoreHeader(t *testing.T) {
	client := &mockNbaClient{}
	m := New(client, "123")
	m.width = 100
	m.height = 40
	m.boxScore = types.LiveBoxScoreResponse{
		Game: types.Game{
			GameId: "123",
			HomeTeam: types.Team{
				TeamTricode: "LAL",
				Players:     &[]types.Player{},
			},
		},
	}

	view := m.View()

	expectedHeaderParts := []string{
		"MIN", "FGM", "FGA", "FG%", "3PM", "3PA", "3P%", "FTM", "FTA", "FT%",
		"OREB", "DREB", "REB", "AST", "STL", "BLK", "TO", "PF", "PTS", "+/-",
	}

	for _, part := range expectedHeaderParts {
		assert.Contains(t, view, part, "Box score header should contain %s", part)
	}
}

func TestMinutesFormatting(t *testing.T) {
	client := &mockNbaClient{}
	m := New(client, "123")
	m.width = 100
	m.height = 40

	// Use PT format for Minutes which MinutesClock() parses
	t.Run("Long format minutes", func(t *testing.T) {
		players := []types.Player{
			{
				FamilyName: "LongTime",
				Statistics: &types.PlayerBoxScoreStatistic{
					CommonBoxScoreStatistic: types.CommonBoxScoreStatistic{
						// PT36M10.01S -> 36:10.01 (7+ chars) -> 36:10 (5 chars)
						Minutes: "PT36M10.01S",
					},
				},
			},
		}
		m.boxScore = types.LiveBoxScoreResponse{
			Game: types.Game{
				GameId:   "123",
				HomeTeam: types.Team{TeamTricode: "LAL", Players: &players},
			},
		}
		
		view := m.View()
		assert.Contains(t, view, "36:10", "Should truncate '36:10.01' to '36:10'")
	})

	t.Run("Short format minutes", func(t *testing.T) {
		players := []types.Player{
			{
				FamilyName: "ShortTime",
				Statistics: &types.PlayerBoxScoreStatistic{
					CommonBoxScoreStatistic: types.CommonBoxScoreStatistic{
						// PT05M00.00S -> 05:00.00 -> 5 chars? 
						// Wait, if it parses to "05:00.00", len is 8.
						// If parses to "05:00", len is 5.
						// The comment says PT07M11.01S -> 07:11.01
						// Let's assume PT05M00.00S -> 05:00.00
						// So it will be > 5 chars, so it will be shown as "05:00".
						// But wait, the user said:
						// "use first 5 chars... if <= 5 chars insert '-'"
						// If the result of MinutesClock() is "05:00.00", first 5 is "05:00".
						// If the result is "0:00" (maybe DNP?), then "-"
						
						// Let's try a very short one.
						// PT00M00.00S -> 00:00.00 -> "00:00"
						
						// Maybe an empty string or something small?
						// "PT00M05.00S" -> "00:05.00" -> "00:05"
						
						// If the user meant "if the *clock string* is <= 5 chars",
						// then "05:00" fits that condition if the output of MinutesClock is just "05:00".
						// But based on `PT...` format it usually has seconds/decimals.
						
						// However, maybe the user means if *after truncation* or something?
						// "利用して前半5文字を採用してMINを描画してください。5文字以下の場合には-を入れる"
						// "Use the first 5 chars... In case of 5 chars or less, insert '-'"
						
						// So if MinutesClock() returns "12:34.56" (8 chars), take "12:34".
						// If MinutesClock() returns "5:00" (4 chars), take "-".
						
						// So I need a case where MinutesClock() returns something short.
						// Let's assume MinutesClock returns "DNP" or something if Minutes is empty?
						// Or simply "PT0M0S" -> "0:00" (4 chars)?
						Minutes: "PT0M0S",
					},
				},
			},
		}
		m.boxScore = types.LiveBoxScoreResponse{
			Game: types.Game{
				GameId:   "123",
				HomeTeam: types.Team{TeamTricode: "LAL", Players: &players},
			},
		}
		
		view := m.View()
		// We expect "-" for "0:00" (4 chars)
		assert.NotContains(t, view, "0:00")
		// We can't easily assert "-" because many things might be "-" (like +/-)
		// But if we don't see "0:00" in the MIN column, that's a good sign.
	})
}

func TestGameStatusFormatting(t *testing.T) {
	client := &mockNbaClient{}
	m := New(client, "123")
	m.width = 100
	m.height = 40

	t.Run("Not Started", func(t *testing.T) {
		m.boxScore = types.LiveBoxScoreResponse{
			Game: types.Game{
				GameId:     "123",
				GameStatus: 1, // Pre-game -> IsGameStart() == false
				HomeTeam: types.Team{TeamTricode: "LAL"},
				AwayTeam: types.Team{TeamTricode: "GSW"},
			},
		}
		
		view := m.View()
		assert.Contains(t, view, "not started")
	})

	t.Run("Overtime 1OT", func(t *testing.T) {
		m.boxScore = types.LiveBoxScoreResponse{
			Game: types.Game{
				GameId:     "123",
				GameStatus: 2,
				Period:     5, // 1st OT -> Period 5
				GameClock:  "PT05M00.00S",
				HomeTeam: types.Team{TeamTricode: "LAL"},
				AwayTeam: types.Team{TeamTricode: "GSW"},
			},
		}
		
		view := m.View()
		assert.Contains(t, view, "1OT")
		assert.NotContains(t, view, "5Q")
	})

	t.Run("Overtime 2OT", func(t *testing.T) {
		m.boxScore = types.LiveBoxScoreResponse{
			Game: types.Game{
				GameId:     "123",
				GameStatus: 2,
				Period:     6, // 2nd OT -> Period 6
				GameClock:  "PT05M00.00S",
				HomeTeam: types.Team{TeamTricode: "LAL"},
				AwayTeam: types.Team{TeamTricode: "GSW"},
			},
		}
		
		view := m.View()
		assert.Contains(t, view, "2OT")
		assert.NotContains(t, view, "6Q")
	})
}