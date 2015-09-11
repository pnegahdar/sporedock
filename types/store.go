package types

import "github.com/pnegahdar/sporedock/utils"

type SubscriptionManager struct {
	ID       string
	Messages chan string
	Exit     utils.SignalCast
}

const SentinelEnd = -1

type SporeStore interface {
	Module
	Get(i interface{}, id string) error
	GetAll(v interface{}, start int, end int) error
	Set(v interface{}, id string, logTrim int) error
	Update(v interface{}, id string, logTrim int) error
	Exists(v interface{}, id string) (bool, error)
	Delete(v interface{}, id string) error
	DeleteAll(v interface{}) error
	Publish(v interface{}, channels ...string) error
	Subscribe(channel string) (*SubscriptionManager, error)
	IsHealthy(sporeName string) (bool, error)
	LeaderName() (string, error)
}
