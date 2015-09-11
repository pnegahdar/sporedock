package types

type Assignment struct {
	SporeID string
	AppIDs  []string
}

func (a *Assignment) GetApp(runContext *RunContext) (*App, error) {
	app := &App{}
	err := runContext.Store.Get(app, a.SporeID)
	return app, err
}

func (a *Assignment) GetSpore(runContext *RunContext) (*Spore, error) {
	spore := &Spore{}
	err := runContext.Store.Get(spore, a.SporeID)
	return spore, err

}
