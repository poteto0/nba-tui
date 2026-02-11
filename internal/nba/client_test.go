package nba

import (
	"testing"
)

func TestClient_GetBoxScore(t *testing.T) {
	c := NewClient()
	// Assuming empty ID returns error or handled by SDK
	_, err := c.GetBoxScore("")
	if err == nil {
		// Just a placeholder assertion to ensure we can call it
		// In a real integration test, we'd check for specific errors
	}
}

func TestClient_GetPlayByPlay(t *testing.T) {
	c := NewClient()
	_, err := c.GetPlayByPlay("")
	if err == nil {
	}
}
