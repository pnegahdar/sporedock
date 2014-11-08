package cluster

type WebApp struct {
	Count        int      `flatten:"{{ .ID }}/Count"`
	Env          string   `flatten:"{{ .ID }}/Env/"`
	ID           string   `flatten:"{{ .ID }}"`
	Image        string   `flatten:"{{ .ID }}/Image"`
	WebEndpoints []string `flatten:"{{ .ID }}/WebEndpoints/"`
	Weight       float32  `flatten:"{{ .ID }}/Weight"`
}

type WebApps []WebApp

// Define the interface for sorting
func (wa WebApps) Len() int           { return len(wa) }
func (wa WebApps) Swap(i, j int)      { wa[i], wa[j] = wa[j], wa[i] }
func (wa WebApps) Less(i, j int) bool { return wa[i].Weight < wa[j].Weight }
