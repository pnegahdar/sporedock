package manifest

type Env struct {
	Env   map[string]string
	EnvID string
}

type WebApp struct {
	AppName        string
	Image          string
	StartupCommand string
	InternalPort   int
	EnvIDs         []string
	RunMax         bool
	RunSingle      bool
	WebEndpoints   []string
}

type Cluster struct {
	Envs    []Env
	WebApps []WebApp
	Name    string
}
