package config

type Env struct {
	Env map[string]string `etcd:"{{ .ID }}/{{ .KEY }}"`
	ID  string            `etcd:"{{ .ID }}/"`
}

type Docker struct {
	Host string `etcd:"Host"`
}

type WebApp struct {
	EnvIDs         []string `etcd:"{{ .ID }}/EnvsIDs"`
	ID             string   `etcd:"{{ .ID }}"`
	Image          string   `etcd:"{{ .ID }}/Image"`
	InternalPort   string   `etcd:"{{ .ID }}/InternalPort"`
	RunMax         string   `etcd:"{{ .ID }}/RunMax"`
	StartupCommand string   `etcd:"{{ .ID }}/StartupCommand"`
	WebEndpoints   []string `etcd:"{{ .ID }}/WebEndpoints"`
}

type Cluster struct {
	Envs    []Env    `etcd:"/sporedock/clusters/{{ .ID }}/envs/"`
	Docker  Docker   `etcd:"/sporedock/clusters/{{ .ID }}/docker/"`
	ID      string   `etcd:"/sporedock/clusters/{{ .ID }}/"`
	WebApps []WebApp `etcd:"/sporedock/cluster/{{ .ID }}/webapps/"`
}
