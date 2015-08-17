package cluster

import (
	"errors"
	"fmt"
	"github.com/pnegahdar/sporedock/types"
	"github.com/pnegahdar/sporedock/utils"
	"sync"
	"sort"
)


const FitRangeBound = 0.05
const PackIfNoFit = true


const currentPlanName = "current"

type appID string
type sporeID string

type Plan struct {
	sync.Mutex
	SporeSchedule map[sporeID][]*App
	AppSchedule   map[appID][]sporeID
	SizeRem       map[sporeID]float64 `json:"-"`
	SporeMap      map[sporeID]*Spore `json:"-"`
	SporesDecr    []Spore `json:"-"`
	setupOnce     sync.Once `json:"-"`
}

func NewPlan(runContext *types.RunContext) (*Plan, error) {
	allSpores, err := AllSpores(runContext)
	sort.Sort(sort.Reverse(allSpores))
	if err == types.ErrNoneFound {
		return nil, err
	}
	utils.HandleError(err)
	sporeMap := map[sporeID]*Spore{}
	for _, spore := range allSpores {
		sporeMap[sporeID(spore.ID)] = &spore
	}
	return &Plan{SporesDecr: allSpores, SporeMap: sporeMap}, nil
}

func (plan *Plan) init() {
	plan.setupOnce.Do(func() {
		plan.SporeSchedule = map[sporeID][]*App{}
		plan.AppSchedule = map[appID][]sporeID{}
		plan.SizeRem = map[sporeID]float64{}
	})
}

func (plan *Plan) Add(spore *Spore, app *App) {
	plan.init()
	plan.SporeSchedule[sporeID(spore.ID)] = append(plan.SporeSchedule[sporeID(spore.ID)], app)
	plan.AppSchedule[appID(app.ID)] = append(plan.AppSchedule[appID(app.ID)], sporeID(spore.ID))
	if _, ok := plan.SizeRem[sporeID(spore.ID)]; !ok {
		plan.SizeRem[sporeID(spore.ID)] = spore.Size()
	}
	plan.SizeRem[sporeID(spore.ID)] = plan.SizeRem[sporeID(spore.ID)] - app.Size()
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
	utils.LogWarnF("Ran into error [%v] when scheduling app %v on %v", err.Error(), appName, fnName)
}

func PinNodeScheduler(app *App, runContext *types.RunContext, currentPlan *Plan, newPlan *Plan) (bool, error) {
	if app.PinSpore != "" {
		spore, ok := newPlan.SporeMap[sporeID(app.PinSpore)]
		if !ok {
			return true, errors.New(fmt.Sprintf("Unable to schedule app %v to %v because spore doesn't exists", app.ID, app.PinSpore))
		}
		for i := 0; i < app.Count; i++ {
			newPlan.Add(spore, app)
		}
		return true, nil
	}
	return false, nil
}

func PersistExistingScheduler(app *App, runContext *types.RunContext, currentPlan *Plan, newPlan *Plan) (bool, error) {
	if spores, ok := currentPlan.AppSchedule[appID(app.ID)]; ok {
		for _, sporename := range (spores) {
			if spore, ok := newPlan.SporeMap[sporeID(sporename)]; ok {
				newPlan.Add(spore, app)
				app.Count--
			}
		}
	}
	if app.Count == 0 {
		return true, nil
	}
	return false, nil
}

func FirstFirstDecreasingScheduler(app *App, runContext *types.RunContext, currentPlan *Plan, newPlan *Plan) error {
	var largestSpore *Spore
	largestRemSize := 0.0
	for i := 0; i < app.Count; i++ {
		for _, spore := range (newPlan.SporesDecr) {
			var size float64
			if _, ok := newPlan.SizeRem[sporeID(spore.ID)]; !ok {
				size = spore.Size()
			}
			if size > largestRemSize || largestSpore == nil {
				largestSpore, largestRemSize = &spore, largestRemSize
			}

			cpuFits := within(app.Cpus, spore.Cpus, FitRangeBound)
			memFits := within(app.Mem, spore.Mem, FitRangeBound)
			if cpuFits && memFits {
				utils.LogInfoF("FITT app %v into %v", app, spore)
				newPlan.Add(&spore, app)
				return nil
			}
		}
		if PackIfNoFit {
			utils.LogWarnF("Failed to fit app %v in any nodes. Forcing somewhere.", app)
			// Puts it in first place with largest rem size
			newPlan.Add(largestSpore, app)
		} else {
			utils.LogWarnF("Failed to fit app %v, packifnofit disabled.", app)
		}
	}
	return nil
}

func within(value float64, base float64, bound float64) bool {
	return value >= 0 && value <= (base * (1.0 + bound))
}
