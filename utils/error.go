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
	errorString := "Ran into the following errors"
	if len(be.errors) > 0 {
		for _, err := range be.errors {
			errorString = errorString + " -- " + err.Error()
		}
		return errors.New(errorString)

	}
	return nil
}
