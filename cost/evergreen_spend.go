package cost

import (
	"time"

	"github.com/evergreen-ci/sink/evergreen"
	"github.com/evergreen-ci/sink/model"
	"github.com/mongodb/grip"
	"github.com/mongodb/grip/message"
	"github.com/pkg/errors"
	"golang.org/x/net/context"
)

const (
	enableEvergreenDistroCollector   = false
	enableEvergreenProjectsCollector = true
)

// GetEvergreenDistrosData returns distros cost data stored in Evergreen by
// calling evergreen GetEvergreenDistrosData function.
func getEvergreenDistrosData(ctx context.Context, c *evergreen.Client, starttime time.Time, duration time.Duration) ([]model.EvergreenDistroCost, error) {
	distros := []model.EvergreenDistroCost{}
	evgDistros, err := c.GetEvergreenDistrosData(ctx, starttime, duration)
	if err != nil {
		return nil, errors.Wrap(err, "error in getting Evergreen distros data")
	}
	for idx := range evgDistros {
		distros = append(distros, convertEvgDistroToCostDistro(evgDistros[idx]))
	}

	return distros, nil
}

func convertEvgDistroToCostDistro(evgdc *evergreen.DistroCost) model.EvergreenDistroCost {
	d := model.EvergreenDistroCost{}
	d.Name = evgdc.DistroID
	d.Provider = evgdc.Provider
	d.InstanceType = evgdc.InstanceType
	d.InstanceSeconds = int64(evgdc.SumTimeTaken / time.Second)
	d.EstimatedCost = evgdc.SumEstimatedCost
	return d
}

// GetEvergreenProjectsData returns distros cost data stored in Evergreen by
// calling evergreen GetEvergreenDistrosData function.
func getEvergreenProjectsData(ctx context.Context, c *evergreen.Client, starttime time.Time, duration time.Duration) ([]model.EvergreenProjectCost, error) {
	projects := []model.EvergreenProjectCost{}
	evgProjects, err := c.GetEvergreenProjectsData(ctx, starttime, duration)
	if err != nil {
		return nil, errors.Wrap(err, "error in getting Evergreen projects data")
	}

	for _, p := range evgProjects {
		projects = append(projects, convertEvgProjectUnitToCostProject(p))
	}

	return projects, nil
}

func convertEvgProjectUnitToCostProject(evgpu evergreen.ProjectUnit) model.EvergreenProjectCost {
	p := model.EvergreenProjectCost{}
	p.Name = evgpu.Name

	grip.Info(message.Fields{
		"message":   "building cost data for evergreen",
		"project":   evgpu.Name,
		"num_tasks": len(evgpu.Tasks),
	})
	for _, task := range evgpu.Tasks {
		costTask := model.EvergreenTaskCost{}
		costTask.Githash = task.Githash
		costTask.Name = task.DisplayName
		costTask.BuildVariant = task.BuildVariant
		costTask.TaskSeconds = int64(task.TimeTaken / time.Second)
		costTask.EstimatedCost = task.Cost
		p.Tasks = append(p.Tasks, costTask)
	}

	return p
}

func getEvergreenData(ctx context.Context, c *evergreen.Client, starttime time.Time, duration time.Duration) (*model.EvergreenCost, error) {
	out := &model.EvergreenCost{}

	if enableEvergreenDistroCollector {
		grip.Info("Getting Evergreen Distros")
		distros, err := getEvergreenDistrosData(ctx, c, starttime, duration)
		if err != nil {
			return nil, errors.Wrap(err, "problem getting distro data from evergreen")
		}
		out.Distros = distros
	}

	if enableEvergreenProjectsCollector {
		grip.Info("Getting Evergreen Projects")
		projects, err := getEvergreenProjectsData(ctx, c, starttime, duration)
		if err != nil {
			return nil, errors.Wrap(err, "problem getting project data from evergreen")
		}
		out.Projects = projects
	}

	return out, nil
}
