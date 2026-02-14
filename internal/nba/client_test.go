package nba

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClient_GetScoreboard(t *testing.T) {
	t.Run("return contents of result", func(t *testing.T) {
		// Arrange
		c := NewClient()

		// Act
		result, err := c.GetScoreboard()

		// Assert
		assert.NoError(t, err)
		assert.NotEmpty(t, result)
	})
}

func TestClient_GetBoxScore(t *testing.T) {
	t.Run("return contents of result", func(t *testing.T) {
		// Arrange
		c := NewClient()

		// Act
		result, err := c.GetBoxScore("0022500733")

		// Assert
		assert.NoError(t, err)
		assert.NotEmpty(t, result)
	})

	t.Run("invalid args is error", func(t *testing.T) {
		// Arrange
		c := NewClient()

		// Act
		_, err := c.GetBoxScore("")

		// Assert
		assert.Error(t, err)
	})
}

func TestClient_GetPlayByPlay(t *testing.T) {
	t.Run("return contents of result", func(t *testing.T) {
		// Arrange
		c := NewClient()

		// Act
		result, err := c.GetPlayByPlay("0022500733")

		// Assert
		assert.NoError(t, err)
		assert.NotEmpty(t, result)
	})

	t.Run("invalid args is error", func(t *testing.T) {
		// Arrange
		c := NewClient()

		// Act
		_, err := c.GetPlayByPlay("")

		// Assert
		assert.Error(t, err)
	})
}
