package config

type Env struct {
	Env map[string]string `etcd:{ .ID }/{ .Key }`
	ID  string            `etcd:{ .ID }`
}

type Docker struct {
	Host string `etcd:/Host`
}

type WebApp struct {
	ID             string   `etcd:{ .ID }`
	Image          string   `etcd:{ .ID }/Image`
	StartupCommand string   `etcd:{ .ID }/StartupCommand`
	InternalPort   int      `etcd:{ .ID }/InternalPort`
	EnvIDs         []string `etcd:{ .ID }/EnvsIDs`
	RunMax         bool     `etcd:{ .ID }/RunMax`
	WebEndpoints   []string `etcd:{ .ID }/WebEndpoints/`
}

type Cluster struct {
	Envs    []Env    `etcd:/sporedock/clusters/{.ID}/envs/`
	WebApps []WebApp `etcd:/sporedock/cluster/{ .ID }/webapps/`
	ID      string   `etcd:/sporedock/clusters/{.ID}`
	Docker  Docker   `etcd:/sporedock/clusters/{.ID}/docker`
}
