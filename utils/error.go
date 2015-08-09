package utils

import (
	"errors"
	"sync"
)

type BatchError struct {
	sync.RWMutex
	errors []error
}

func (be *BatchError) Add(err error) {
	be.Lock()
	defer be.Unlock()
	if err != nil {
		be.errors = append(be.errors, err)
	}
}

func (be *BatchError) Error() error {
	be.RLock()
	defer be.RUnlock()
	errorString := ""
	if len(be.errors) > 0 {
		for _, err := range be.errors {
			errorString = errorString + "--" + err.Error() + "\n"
		}
		return errors.New(errorString)

	}
	return nil
}
