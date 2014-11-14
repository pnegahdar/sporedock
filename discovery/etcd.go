package discovery

import (
	"errors"
	"fmt"
	"github.com/pnegahdar/sporedock/server"
	"github.com/pnegahdar/sporedock/settings"
	"github.com/pnegahdar/sporedock/utils"
	"strings"
	"net/url"
)

type Machine struct {
	Name      string
	State     string
	ClientURL string
	PeerURL   string
}

func ListMachines() []Machine {
	resp, err := server.EtcdPeerClient().GetMachines("http://127.0.0.1:7001")
	if err != nil {
		utils.HandleError(errors.New(err.Message))
	}
	machines := []Machine{}
	for _, m := range resp {
		machines = append(machines, Machine{Name: m.Name, State: m.State, ClientURL: m.ClientURL, PeerURL: m.PeerURL})
	}
	return machines
}

func CurrentMachine() Machine {
	machines := ListMachines()
	for _, v := range machines {
		if strings.Index(v.Name, settings.GetInstanceName()) != -1 {
			return v
		}
	}
	utils.HandleError(errors.New("Current machine not found!"))
	return Machine{}
}

func GetLeader() (Machine, error) {
	machines := ListMachines()
	for _, v := range machines {
		if v.State == "leader" {
			return v, nil
		}
	}
	err := errors.New("Leader not found")
	return Machine{}, err
}
func AmLeader() bool {
	me := CurrentMachine()
	return me.State == "leader"
}

func (m Machine) GetIP() string {
	u, err := url.Parse(m.PeerURL)
	utils.HandleError(err)
	return strings.Split(u.Host, ":")[0]
}

func (m Machine) GetPortLocation(port string) string {
	return fmt.Sprintf("http://%v:%v", m.GetIP(), port)
}
