package config

type Env struct {
	Env map[string]string `flatten:"{{ .ID }}/{{ .KEY }}"`
	ID  string            `flatten:"{{ .ID }}/"`
}

type Docker struct {
	Host string `flatten:"Host"`
}

type WebApp struct {
	EnvIDs         []string `flatten:"{{ .ID }}/EnvsIDs/"`
	ID             string   `flatten:"{{ .ID }}"`
	Image          string   `flatten:"{{ .ID }}/Image"`
	InternalPort   string   `flatten:"{{ .ID }}/InternalPort"`
	RunMax         string   `flatten:"{{ .ID }}/RunMax"`
	StartupCommand string   `flatten:"{{ .ID }}/StartupCommand"`
	WebEndpoints   []string `flatten:"{{ .ID }}/WebEndpoints/"`
}

type Cluster struct {
	Envs    []Env    `flatten:"/sporedock/clusters/{{ .ID }}/envs/"`
	Docker  Docker   `flatten:"/sporedock/clusters/{{ .ID }}/docker/"`
	ID      string   `flatten:"/sporedock/clusters/{{ .ID }}/"`
	WebApps []WebApp `flatten:"/sporedock/cluster/{{ .ID }}/webapps/"`
}
