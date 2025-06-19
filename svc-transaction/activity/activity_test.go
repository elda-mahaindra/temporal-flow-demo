package activity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetActivities(t *testing.T) {
	activity := &Activity{}
	activities := activity.GetActivities()

	// Should have exactly 3 activities
	assert.Equal(t, 3, len(activities))

	// All activities should be non-nil
	for _, act := range activities {
		assert.NotNil(t, act)
	}
}
