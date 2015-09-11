package types

import (
	"errors"
	"fmt"
	"github.com/pnegahdar/sporedock/utils"
	"math"
	"sort"
	"sync"
)

const FitRangeBound = 0.05
const PackIfNoFit = true

const currentPlanName = "current"

type RunGuid string

type SporeGuid struct {
	Sporeid SporeID
	Appguid RunGuid
}

type Plan struct {
	sync.Mutex
	SporeSchedule map[SporeID]map[RunGuid]*App
	AppSchedule   map[AppID][]SporeGuid
	SizeRem       map[SporeID]float64 `json:"-"`
	SporeMap      map[SporeID]*Spore  `json:"-"`
	SporesDecr    []Spore             `json:"-"`
	setupOnce     sync.Once           `json:"-"`
}

func NewPlan(runContext *RunContext) (*Plan, error) {
	// Todo: exclude watchers
	allSpores, err := AllSpores(runContext)
	sort.Sort(sort.Reverse(allSpores))
	if err == ErrNoneFound {
		return nil, err
	}
	utils.HandleError(err)
	sporeMap := map[SporeID]*Spore{}
	for _, spore := range allSpores {
		sporeMap[SporeID(spore.ID)] = &spore
	}
	return &Plan{SporesDecr: allSpores, SporeMap: sporeMap}, nil
}

func (plan *Plan) init() {
	plan.setupOnce.Do(func() {
		plan.SporeSchedule = map[SporeID]map[RunGuid]*App{}
		plan.AppSchedule = map[AppID][]SporeGuid{}
		plan.SizeRem = map[SporeID]float64{}
	})
}

func (plan *Plan) Add(spore *Spore, app *App, guid RunGuid) {
	plan.init()
	if _, ok := plan.SporeSchedule[SporeID(spore.ID)]; !ok {
		plan.SporeSchedule[SporeID(spore.ID)] = map[RunGuid]*App{}
	}
	plan.SporeSchedule[SporeID(spore.ID)][guid] = app
	plan.AppSchedule[AppID(app.ID)] = append(plan.AppSchedule[AppID(app.ID)], SporeGuid{Sporeid: SporeID(spore.ID), Appguid: guid})
	if _, ok := plan.SizeRem[SporeID(spore.ID)]; !ok {
		plan.SizeRem[SporeID(spore.ID)] = spore.Size()
	}
	plan.SizeRem[SporeID(spore.ID)] = plan.SizeRem[SporeID(spore.ID)] - app.Size()
}

func CurrentPlan(runContext *RunContext) (*Plan, error) {
	plan := &Plan{}
	err := runContext.Store.Get(plan, currentPlanName)
	if err != nil {
		return nil, err
	}
	return plan, nil
}

func SavePlan(runContext *RunContext, plan *Plan) error {
	err := runContext.Store.Update(plan, currentPlanName, SentinelEnd)
	return err
}

type schedulerFunc func(app *App, runContext *RunContext, currentPlan *Plan, newPlan *Plan) (bool, error)

var Schedulers = []schedulerFunc{PinNodeScheduler, PersistExistingScheduler}
var FinalScheduler = FirstFitsDecreasingScheduler

func HandleSchedulerError(err error, appName AppID, fnName string) {
	utils.LogWarnF("Ran into error [%v] when scheduling app %v on %v", err.Error(), appName, fnName)
}

func PinNodeScheduler(app *App, runContext *RunContext, currentPlan *Plan, newPlan *Plan) (bool, error) {
	if app.PinSpore != "" {
		spore, ok := newPlan.SporeMap[SporeID(app.PinSpore)]
		if !ok {
			return true, errors.New(fmt.Sprintf("Unable to schedule app %v to %v because spore doesn't exists", app.ID, app.PinSpore))
		}
		for i := 0; i < app.Count; i++ {
			newPlan.Add(spore, app, RunGuid(utils.GenGuid()))
		}
		return true, nil
	}
	return false, nil
}

func PersistExistingScheduler(app *App, runContext *RunContext, currentPlan *Plan, newPlan *Plan) (bool, error) {
	if sporeguids, ok := currentPlan.AppSchedule[AppID(app.ID)]; ok {
		packCount := int(math.Min(float64(app.CountRemaining), float64(len(sporeguids))))
		for i := 0; i < packCount; i++ {
			if spore, ok := newPlan.SporeMap[sporeguids[i].Sporeid]; ok {
				newPlan.Add(spore, app, sporeguids[i].Appguid)
				app.CountRemaining--
			}
		}
	}
	if app.CountRemaining == 0 {
		return true, nil
	}
	return false, nil
}

func FirstFitsDecreasingScheduler(app *App, runContext *RunContext, currentPlan *Plan, newPlan *Plan) error {
	var largestSpore *Spore
	largestRemSize := 0.0
	for i := 0; i < app.CountRemaining; i++ {
		for _, spore := range newPlan.SporesDecr {
			var size float64
			if _, ok := newPlan.SizeRem[SporeID(spore.ID)]; !ok {
				size = spore.Size()
			}
			if size > largestRemSize || largestSpore == nil {
				largestSpore, largestRemSize = &spore, largestRemSize
			}

			cpuFits := within(app.Cpus, spore.Cpus, FitRangeBound)
			memFits := within(app.Mem, spore.Mem, FitRangeBound)
			if cpuFits && memFits {
				utils.LogInfoF("FITT app %v into %v", app, spore)
				newPlan.Add(&spore, app, RunGuid(utils.GenGuid()))
				return nil
			}
		}
		if PackIfNoFit {
			utils.LogWarnF("Failed to fit app %v in any nodes. Forcing somewhere.", app)
			// Puts it in first place with largest rem size
			newPlan.Add(largestSpore, app, RunGuid(utils.GenGuid()))
		} else {
			utils.LogWarnF("Failed to fit app %v, packifnofit disabled.", app)
		}
	}
	return nil
}

func within(value float64, base float64, bound float64) bool {
	return value >= 0 && value <= (base*(1.0+bound))
}
