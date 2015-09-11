package types

const HostMapLogLength = 200

type HostMap struct {
	Host     string
	AppNames []string
}

func (hm HostMap) Validate() error {
	return nil
}
