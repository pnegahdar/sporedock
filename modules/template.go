package modules
import (
	"sync"
	"github.com/pnegahdar/sporedock/utils"
	"github.com/pnegahdar/sporedock/types"
	"time"
	"fmt"
)

// Tempalte for a new module.
// Replace:
// 		Template: ModuleName
// 		tmpl : Module shortname
type Template struct {
	initOnce   sync.Once
	stopCast   utils.SignalCast
	runContext *types.RunContext
}

func (tmpl *Template) Init(runContext *types.RunContext) {
	tmpl.initOnce.Do(func() {
		tmpl.runContext = runContext
	})
}

func (tmpl *Template) ProcName() string {
	return "Template"
}

func (tmpl *Template) Stop() {
	tmpl.stopCast.Signal()
}

func (tmpl *Template) Run(runContext *types.RunContext) {
	exit, _ := tmpl.stopCast.Listen()
	for {
		select {
		case <-time.After(time.Second * 10):
			fmt.Println("Running from ", tmpl.ProcName())
		case <-exit:
			return
		}
	}
}


func (tmpl *Template) ShouldRun(runContext *types.RunContext) bool{
	return true
}
