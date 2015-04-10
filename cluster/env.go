package cluster

import (
    "fmt"
    "github.com/pnegahdar/sporedock/utils"
    "github.com/pnegahdar/sporedock/store"
)

type Env struct {
    Env map[string]string
    ID  string
}

type Envs []Env

func (e Env) AsDockerSlice() []string {
    return EnvAsDockerKV(e.Env)
}

func (e Env) TypeIdentifier() string {
    return "env"
}

func (e Env) MyIdentifier() string {
    return e.ID
}

func (e Env) ToString() string {
    return utils.Marshall(e)
}

func (e Env) validate() error {
    return nil
}

func (e *Env) FromString(data string) (*store.Storable, error) {
    utils.Unmarshall(data, e)
    err := e.validate()
    return e, err
}

func FindEnv(name string) Env {
    // TODO
    return Env
}
func EnvAsDockerKV(envVars map[string]string) []string {
    data := []string{}
    for k, v := range envVars {
        data = append(data, fmt.Sprintf("%v=%v", k, v))
    }
    return data
}

func FlattenEnvs(envs ...Env) map[string]string {
    finalMap := map[string]string{}
    for _, env := range (envs) {
        for k, v := range (env.Env) {
            finalMap[k] = v
        }
    }
}
