package cluster

import (
	"errors"
	"fmt"
	"github.com/pnegahdar/sporedock/types"
	"github.com/pnegahdar/sporedock/utils"
)

type Schedule struct {
	Spore *Spore
	apps  []*App
}

func (s *Schedule) AddApp(app *App) {
	s.apps = append(s.apps, app)
}

type Schedules []*Schedule

const currentPlanName = "current"

type Plan struct {
	SporeSchedule map[string]*Schedule
	Schedules     Schedules         `json:"-"`
	Spores        map[string]*Spore `json:"-"`
}

func (plan *Plan) Assign(sporeID string, app *App) {
	schedule, ok := plan.SporeSchedule[sporeID]
	if schedule == nil || !ok {
		schedule := &Schedule{Spore: plan.Spores[sporeID], apps: []*App{}}
		plan.SporeSchedule[sporeID] = schedule
		plan.Schedules = append(plan.Schedules, schedule)
	}
	plan.SporeSchedule[sporeID].AddApp(app)
	plan.Schedules = append(plan.Schedules, schedule)

}

func CurrentPlan(runContext *types.RunContext) (*Plan, error) {
	plan := &Plan{}
	err := runContext.Store.Get(plan, currentPlanName)
	if err != nil {
		return nil, err
	}
	return plan, nil
}

func SavePlan(runContext *types.RunContext, plan *Plan) error {
	err := runContext.Store.Update(plan, currentPlanName, types.SentinelEnd)
	return err
}

type schedulerFunc func(app *App, runContext *types.RunContext, currentPlan *Plan, newPlan *Plan) (bool, error)

var Schedulers = []schedulerFunc{PinNodeScheduler, PersistExistingScheduler}
var FinalScheduler = FirstFirstDecreasingScheduler

func HandleSchedulerError(err error, appName string, fnName string) {
	utils.LogWarn(fmt.Sprintf("Ran into error [%v] when scheduling app %v on %v", err.Error(), appName, fnName))
}

func PinNodeScheduler(app *App, runContext *types.RunContext, currentPlan *Plan, newPlan *Plan) (bool, error) {
	if app.PinSpore != "" {
		_, ok := newPlan.Spores[app.PinSpore]
		if !ok {
			return true, errors.New(fmt.Sprintf("Unable to schedule app %v to %v because spore doesn't exists", app.ID, app.PinSpore))
		}
		for i := 0; i < app.Count; i++ {
			newPlan.Assign(app.PinSpore, app)
		}
		return true, nil
	}
	return false, nil
}

func PersistExistingScheduler(app *App, runContext *types.RunContext, currentPlan *Plan, newPlan *Plan) (bool, error) {
	return false, nil
}

func FirstFirstDecreasingScheduler(app *App, runContext *types.RunContext, currentPlan *Plan, newPlan *Plan) error {
	return nil
}
