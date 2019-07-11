package dcos

import (
	"testing"

	marathon "github.com/gambol99/go-marathon"
	"github.com/stretchr/testify/assert"
)

func TestMarathonApplicationStatus(t *testing.T) {
	app := marathon.NewDockerApplication()
	app.Deployments = make([]map[string]string, 1)

	assert.Equal(t, "Deploying", marathonApplicationStatus(app))
	assert.Equal(t, "Unknown", marathonApplicationStatus(nil))
	assert.Equal(t, "Unknown", marathonApplicationStatus(marathon.NewDockerApplication()))
	suspendedApp := marathon.NewDockerApplication().Count(0)
	assert.Equal(t, "Suspended", marathonApplicationStatus(suspendedApp))

	runningApp := marathon.NewDockerApplication().Count(1)
	runningApp.TasksRunning = 1
	assert.Equal(t, "Running", marathonApplicationStatus(runningApp))
}
