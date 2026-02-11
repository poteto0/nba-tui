package nba

import (
	"github.com/poteto0/go-nba-sdk/gns"
	"github.com/poteto0/go-nba-sdk/types"
)

type Client struct {
	gnsClient *gns.Client
}

func NewClient() *Client {
	return &Client{
		gnsClient: gns.NewClient(nil),
	}
}

func (c *Client) GetScoreboard() ([]types.Game, error) {
	result := c.gnsClient.Live.GetScoreBoard(nil)
	if result.Error != nil {
		return nil, result.Error
	}
	return result.Contents.Scoreboard.Games, nil
}

func (c *Client) GetBoxScore(gameID string) (types.LiveBoxScoreResponse, error) {
	result := c.gnsClient.Live.GetBoxScore(&types.BoxScoreParams{GameID: gameID})
	if result.Error != nil {
		return types.LiveBoxScoreResponse{}, result.Error
	}
	return result.Contents, nil
}

func (c *Client) GetPlayByPlay(gameID string) (types.LivePlayByPlayResponse, error) {
	result := c.gnsClient.Live.GetPlayByPlay(&types.PlayByPlayParams{GameID: gameID})
	if result.Error != nil {
		return types.LivePlayByPlayResponse{}, result.Error
	}
	return result.Contents, nil
}
