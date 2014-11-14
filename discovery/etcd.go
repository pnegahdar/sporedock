package discovery

import (
	"errors"
	"fmt"
	"github.com/pnegahdar/sporedock/server"
	"github.com/pnegahdar/sporedock/settings"
	"github.com/pnegahdar/sporedock/utils"
	"net/url"
	"strings"
)

type Machine struct {
	Name      string
	State     string
	ClientURL string
	PeerURL   string
}

func getAPeerUrl(discoveryUrl string) string{
	server.EtcdClient().SyncCluster()
	cluster := server.EtcdClient().GetCluster()
	parsedCluster, err := url.Parse(cluster[0])
	utils.HandleError(err)
	ip := strings.Split(parsedCluster.Host, ":")[0]
	peerUrl := fmt.Sprintf("http://%v:7001", ip)
	return peerUrl
}

func getMachinesFromPeerUrl(peerUrl string) []Machine{
	return []Machine{}
//	resp, err := http.Get(peerUrl + "/v2/admin/machines")
//	utils.HandleError(err)
//	defer resp.Body.Close()
//	body, err := ioutil.ReadAll(resp.Body)
//	utils.HandleError(err)
//	var data etcdclient.Response
//	err = json.Unmarshal(body, &data)
//	utils.HandleError(err)
//	machines := []Machine{}
//	for _, m := range resp {
//		machines = append(machines, Machine{Name: m.Name, State: m.State, ClientURL: m.ClientURL, PeerURL: m.PeerURL})
//	}
//	return machines
}

func ListMachines() []Machine {
	machines := getMachinesFromPeerUrl(getAPeerUrl(settings.GetDiscoveryString()))
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
