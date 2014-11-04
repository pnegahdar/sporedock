package config

import (
	"bytes"
	"encoding/json"
	"github.com/pnegahdar/sporedock/utils"
)

func indentJSon(marshalled []byte) string {
	var buffer bytes.Buffer
	err := json.Indent(&buffer, marshalled, "", "    ")
	utils.HandleError(err)
	return buffer.String()
}
