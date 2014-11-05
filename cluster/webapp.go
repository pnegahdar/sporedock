package cluster

type WebApp struct {
	Env          string   `flatten:"{{ .ID }}/Env/"`
	ID           string   `flatten:"{{ .ID }}"`
	Image        string   `flatten:"{{ .ID }}/Image"`
	WebEndpoints []string `flatten:"{{ .ID }}/WebEndpoints/"`
}
type WebApps []WebApp

