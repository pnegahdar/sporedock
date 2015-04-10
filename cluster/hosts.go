package cluster

import (
    "github.com/pnegahdar/sporedock/utils"
    "github.com/pnegahdar/sporedock/store"
)

const HostMapLogLength = 200

type HostMap struct {
    Host string
    AppName string
}

func (hm HostMap) Identifier() string {
    return hm.Host
}

func (hm HostMap) TypeIdentifier() string {
    return "host"
}

func (hm HostMap) ToString() string {
    return utils.Marshall(hm)
}

func (hm HostMap) validate() error {
    return nil
}

func (hm *HostMap) FromString(data string) (*store.Storable, error) {
    utils.Unmarshall(data, hm)
    err := hm.validate()
    return hm, err
}

