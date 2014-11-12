package cluster

import "fmt"

type Env struct {
	Env map[string]string `flatten:"{{ .ID }}/{{ .KEY }}"`
	ID  string            `flatten:"{{ .ID }}/"`
}

type Envs []Env

func (e Env) AsDockerSlice() []string {
	data := []string{}
	for k, v := range e.Env {
		data = append(data, fmt.Sprintf("%v=%v", k, v))
	}
	return data
}
