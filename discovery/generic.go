package discovery

import (
	"errors"
	"fmt"
	"github.com/docker/docker/vendor/src/github.com/docker/libcontainer/user"
	"net"
	"strings"
)

type MemberType int

const (
	TypeMemberLeader MemberType = iota
	TypeMember
	TypeWatcher
)

const ConnectionStringError = errors.New("Connection string must start with redis://")
const ConnectionStringNotSetError = errors.New("Connection string not set.")

type Member struct {
	Group      string
	MemberIP   net.IP
	MemberType MemberType
}

type SporeStore interface {
	ListMembers() []Member
	GetLeader() Member
	GetMe() Member
	AmLeader() bool
	Run(group string, myType MemberType, myIP net.IP)
}

func GetorCreateStore(connectionString string) (*SporeStore, error) {
	// Return redis store
	if strings.HasPrefix(connectionString, "redis://") {
		return &RedisStore{connectionString}
	} else {
		return nil, ConnectionStringError
	}
}
