package cluster

type Env struct {
	Env map[string]string `flatten:"{{ .ID }}/{{ .KEY }}"`
	ID  string            `flatten:"{{ .ID }}/"`
}

type Envs []Env
