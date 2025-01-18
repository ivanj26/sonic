package redis

import (
	"ivanj26/sonic/constant/state"

	"github.com/go-redis/redis/v8"
)

type IRedis interface {
	Close() error
	GetClusterMyId() (string, error)
	GetClusterSlots() (*ClusterSlotResult, error)
	GetAddr() string
	GetKeysInSlot(slot string, limitCount int) []string
	IsMaster() bool
	SetSlot(slot string, state state.State, nodeId string) error
	Reshard(slot []string, destCli IRedis, maxConcurrent int) error
}

type Redis struct {
	cli *redis.Client
}

const (
	OK      string = "OK"
	BUSYKEY string = "BUSYKEY"
	NOKEY   string = "NOKEY"
)
