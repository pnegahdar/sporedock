package cluster

type WorkerApp struct {
	Count  int     `flatten:"{{ .ID }}/Count"`
	Env    string  `flatten:"{{ .ID }}/Env/"`
	ID     string  `flatten:"{{ .ID }}"`
	Image  string  `flatten:"{{ .ID }}/Image"`
	Weight float32 `flatten:"{{ .ID }}/Weight"`
}

type WorkerApps []WorkerApp

// Define the interface for sorting
func (wa WorkerApps) Len() int           { return len(wa) }
func (wa WorkerApps) Swap(i, j int)      { wa[i], wa[j] = wa[j], wa[i] }
func (wa WorkerApps) Less(i, j int) bool { return wa[i].Weight < wa[j].Weight }

