package app

import (
	"fmt"
	"github.com/pnegahdar/sporedock/settings"
)

func Initialize(discovery string) {
	settings.SetDiscoveryString(discovery)
	fmt.Println("Discovery url successfully set to", discovery)
}
