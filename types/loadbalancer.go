package types

import (
	"fmt"
	"github.com/pnegahdar/sporedock/utils"
)

type LoadBalancerManager struct {
}

type Hostname string

type LoadBalanced struct {
	Hostname
	ID AppID
}

type AppHost struct {
	ID   Hostname
	Apps []AppID
}

func GetHostMap(runContext *RunContext) map[Hostname][]string {
	mapping := map[Hostname][]string{}
	appHosts := []AppHost{}
	runContext.Store.GetAll(appHosts, 0, SentinelEnd)
	currentPlan, err := CurrentPlan(runContext)
	utils.HandleError(err)
	for _, apphost := range appHosts {
		for _, app := range apphost.Apps {
			for _, appguid := range currentPlan.AppSchedule[app] {
				if runtime, ok := currentPlan.SporeSchedule[appguid.Sporeid]; ok {
					for runguid, app := range runtime {
						spore, err := GetSpore(runContext, appguid.Sporeid)
						utils.HandleError(err)
						var host string
						if spore.ID == runContext.MyMachineID {
							port := GetPortOn(runContext, spore, app, runguid)
							host = fmt.Sprintf("http://127.0.0.1:%v", port)
						} else {
							host = fmt.Sprintf("http://%v:80")
						}
						if _, ok := mapping[apphost.ID]; !ok {
							mapping[apphost.ID] = []string{}
						}
						mapping[apphost.ID] = append(mapping[apphost.ID], host)
					}
				}
			}

		}

	}
	return mapping
}
