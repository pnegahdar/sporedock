package cluster

const HostMapLogLength = 200

type HostMap struct {
	Host    string
	AppName string
}

func (hm HostMap) Validate() error {
	return nil
}
