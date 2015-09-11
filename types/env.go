package types

import (
	"fmt"
)

type Env struct {
	Env map[string]string
	ID  string
}

type Envs []Env

func (e Env) AsDockerSlice() []string {
	return EnvAsDockerKV(e.Env)
}

func FindEnv(name string) Env {
	// TODO
	return Env{}
}

func EnvAsDockerKV(envVars map[string]string) []string {
	data := []string{}
	for k, v := range envVars {
		data = append(data, fmt.Sprintf("%v=%v", k, v))
	}
	return data
}
