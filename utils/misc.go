package utils

import "github.com/satori/go.uuid"

func GenGuid() string {
	uuid := uuid.NewV4()
	return uuid.String()
}
