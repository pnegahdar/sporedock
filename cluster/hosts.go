package cluster

import (
	"github.com/pnegahdar/sporedock/utils"
)

const HostMapLogLength = 200

type HostMap struct {
	Host    string
	AppName string
}

func (hm HostMap) Identifier() string {
	return hm.Host
}

func (hm HostMap) TypeIdentifier() string {
	return "host"
}

func (hm HostMap) ToString() string {
	data, err := utils.Marshall(hm)
	utils.HandleError(err)
	return data
}

func (hm HostMap) validate() error {
	return nil
}

func (hm HostMap) FromString(data string) (HostMap, error) {
	hm = HostMap{}
	utils.Unmarshall(data, &hm)
	err := hm.validate()
	return hm, err
}
