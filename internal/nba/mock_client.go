package nba

import (
	"github.com/poteto0/go-nba-sdk/types"
)

type MockClient struct{}

func NewMockClient() *MockClient {
	return &MockClient{}
}

func (c *MockClient) GetScoreboard() ([]types.Game, error) {
	return []types.Game{
		{
			GameId:         "0012300001",
			GameStatus:     2, // Live
			GameStatusText: "Q4 2:00",
			Period:         4,
			GameClock:      "PT02M00.00S",
			HomeTeam: types.Team{
				TeamName:    "Lakers",
				TeamTricode: "LAL",
				Score:       102,
			},
			AwayTeam: types.Team{
				TeamName:    "Warriors",
				TeamTricode: "GSW",
				Score:       99,
			},
		},
		{
			GameId:         "0012300002",
			GameStatus:     3, // Final
			GameStatusText: "Final",
			HomeTeam: types.Team{
				TeamName:    "Celtics",
				TeamTricode: "BOS",
				Score:       110,
			},
			AwayTeam: types.Team{
				TeamName:    "Heat",
				TeamTricode: "MIA",
				Score:       105,
			},
		},
	}, nil
}

func (c *MockClient) GetBoxScore(gameID string) (types.LiveBoxScoreResponse, error) {
	min := "PT35M00.00S"
	pts := 30
	reb := 10
	ast := 8

	pInt := func(i int) *int { return &i }
	pFloat := func(f float64) *float64 { return &f }

	homePlayers := []types.Player{
		{
			FirstName:  "LeBron",
			FamilyName: "James",
			PersonID:   2544,
			Statistics: &types.PlayerBoxScoreStatistic{
				CommonBoxScoreStatistic: types.CommonBoxScoreStatistic{
					Minutes: min,
					Pts:     pInt(pts),
					Reb:     pInt(reb),
					Ast:     pInt(ast),
					FgM:     pInt(10),
					FgA:     pInt(20),
					FgPct:   pFloat(0.5),
					Fg3M:    pInt(2),
					Fg3A:    pInt(5),
					Fg3Pct:  pFloat(0.4),
					FtM:     pInt(8),
					FtA:     pInt(10),
					FtPct:   pFloat(0.8),
					OReb:    pInt(2),
					DReb:    pInt(8),
					Stl:     pInt(2),
					Blk:     pInt(1),
					Tov:     pInt(3),
					PF:      pInt(2),
				},
				PlusMinus: pFloat(5),
			},
		},
	}

	awayPlayers := []types.Player{
		{
			FirstName:  "Stephen",
			FamilyName: "Curry",
			PersonID:   201939,
			Statistics: &types.PlayerBoxScoreStatistic{
				CommonBoxScoreStatistic: types.CommonBoxScoreStatistic{
					Minutes: "PT34M00.00S",
					Pts:     pInt(28),
					Reb:     pInt(5),
					Ast:     pInt(6),
				},
				PlusMinus: pFloat(-2),
			},
		},
	}

	return types.LiveBoxScoreResponse{
		Game: types.Game{
			GameId: gameID,
			HomeTeam: types.Team{
				TeamId:      1,
				TeamTricode: "LAL",
				Score:       110,
				Players:     &homePlayers,
				Statistics: &types.TeamBoxScoreStatistic{
					CommonBoxScoreStatistic: types.CommonBoxScoreStatistic{
						Pts:     pInt(110),
						Reb:     pInt(45),
						Ast:     pInt(25),
						Minutes: "PT240M00.00S",
					},
				},
			},
			AwayTeam: types.Team{
				TeamId:      2,
				TeamTricode: "GSW",
				Score:       100,
				Players:     &awayPlayers,
				Statistics: &types.TeamBoxScoreStatistic{
					CommonBoxScoreStatistic: types.CommonBoxScoreStatistic{
						Pts:     pInt(100),
						Reb:     pInt(40),
						Ast:     pInt(20),
						Minutes: "PT240M00.00S",
					},
				},
			},
		},
	}, nil
}

func (c *MockClient) GetPlayByPlay(gameID string) (types.LivePlayByPlayResponse, error) {
	return types.LivePlayByPlayResponse{
		Game: types.PlayByPlayGame{
			GameID: gameID,
			Actions: []types.Action{
				{
					ActionNumber: 1,
					Clock:        "11:00",
					Period:       1,
					TeamID:       1610612747, // LAL
					Description:  "Jump Ball James vs Curry",
				},
				{
					ActionNumber: 2,
					Clock:        "10:45",
					Period:       1,
					TeamID:       1610612747, // LAL
					Description:  "James 2pt Shot Made",
				},
				{
					ActionNumber: 3,
					Clock:        "10:30",
					Period:       1,
					TeamID:       1610612744, // GSW
					Description:  "Curry 3pt Shot Missed",
				},
			},
		},
	}, nil
}
