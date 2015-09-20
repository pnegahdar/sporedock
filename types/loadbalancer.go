package types

import (
	"fmt"
	"github.com/pnegahdar/sporedock/utils"
)

type LoadBalancerManager struct {
}

func GetMyHostMap(runContext *RunContext) map[Hostname]map[string]bool {
	mapping := map[Hostname]map[string]bool{}
	appHosts := []AppHost{}
	err := runContext.Store.GetAll(&appHosts, 0, SentinelEnd)
	if len(appHosts) == 0 {
		return mapping
	}
	utils.HandleError(err)
	currentPlan, err := CurrentPlan(runContext)
	if err == ErrNoneFound {
		return mapping
	}
	utils.HandleError(err)

	for _, apphost := range appHosts {
		for _, appIDNeedRouting := range apphost.Apps {
			for _, sporeguid := range currentPlan.AppSchedule[appIDNeedRouting] {
				if runguids, ok := currentPlan.SporeSchedule[sporeguid.Sporeid]; ok {
					for runguid, runapp := range runguids {
						if runapp.ID != appIDNeedRouting {
							continue
						}
						spore, err := GetSpore(runContext, sporeguid.Sporeid)
						utils.HandleError(err)
						var host string
						if spore.ID == runContext.MyMachineID {
							port := GetPortOn(runContext, spore, runapp, runguid)
							if port == "0" {
								utils.LogWarnF("Unable to loadbalance app %+v.", runapp)
							}
							host = fmt.Sprintf("http://127.0.0.1:%v", port)
						} else {
							// Todo(parham): check spores bound http port
							host = fmt.Sprintf("http://%v:80")
						}
						if _, ok := mapping[apphost.ID]; !ok {
							mapping[apphost.ID] = map[string]bool{}
						}
						mapping[apphost.ID][host] = true
					}
				}
			}

		}

	}
	return mapping
}
