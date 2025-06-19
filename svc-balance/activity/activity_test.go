package activity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRegisterActivities(t *testing.T) {
	// Create a mock API instance
	api := &Activity{}

	activities := api.GetActivities()

	assert.NotNil(t, activities)
	assert.Len(t, activities, 1, "Expected exactly 1 activity to be registered")
}
