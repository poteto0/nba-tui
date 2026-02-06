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
		gnsClient: gns.NewClient(),
	}
}

func (c *Client) GetScoreboard() ([]types.Game, error) {
	result := c.gnsClient.Live.GetScoreBoard(nil)
	if result.Error != nil {
		return nil, result.Error
	}
	return result.Contents.Scoreboard.Games, nil
}
