package types

type Hostname string

type AppHost struct {
	ID   Hostname
	Apps []AppID
}

func (ah *AppHost) Validate(rc *RunContext) error {
	// TODO(parham): Verify that app exists
	return nil
}

func (ah *AppHost) GetID() string {
	return string(ah.ID)
}
