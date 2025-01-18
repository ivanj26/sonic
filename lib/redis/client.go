package redis

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"ivanj26/sonic/constant"
	"ivanj26/sonic/constant/command"
	"ivanj26/sonic/constant/state"
	"ivanj26/sonic/util/logger"
	"ivanj26/sonic/util/parser"

	"github.com/go-redis/redis/v8"
)

func NewRedisClient(opt *redis.Options) IRedis {
	cli := redis.NewClient(opt)
	if cli == nil {
		logger.Fatal("Failed to initialize the goredis client")
	}

	return &Redis{
		cli: cli,
	}
}

func (r *Redis) Close() error {
	return r.cli.Close()
}

func (r *Redis) GetClusterMyId() (string, error) {
	id, err := r.cli.Do(context.Background(), command.CLUSTER_MYID...).Result()
	if err != nil {
		logger.Errorf("Failed to get 'cluster myid' node. IP: %s, Err=%s", r.cli.Options().Addr, err)
		return "", err
	}
	idStr, ok := id.(string)
	if !ok {
		logger.Errorf("Failed to get 'cluster myid' node. IP: %s, reason= unknown value", r.cli.Options().Addr)
		return "", errors.New("cluster MYID type result is not string")
	}

	return idStr, nil
}

func (r *Redis) GetClusterSlots() (*ClusterSlotResult, error) {
	res, err := r.cli.Do(context.Background(), command.CLUSTER_SLOTS...).Result()
	if err != nil {
		logger.Errorf("Failed to get `cluster slots` on node ip: %s. Err=%s", r.GetAddr(), err)
		return nil, err
	}

	resT, ok := res.([]interface{})
	if !ok {
		logger.Errorf("Failed to parse `cluster slots` result on node ip: %s", r.GetAddr())
		return nil, err
	}

	allSlots, filteredSlots := (*ClusterSlotResult)(&resT), make([]interface{}, 0)
	myIp := strings.Split(r.GetAddr(), ":")[0]
	if allSlots != nil {
		for _, u := range *allSlots {
			nestedU := u.([]interface{})
			if len(nestedU) >= 2 {
				ipInfo := nestedU[2].([]interface{})
				if len(ipInfo) > 0 && ipInfo[0] == myIp {
					filteredSlots = append(filteredSlots, nestedU)
				}
			}
		}

		return (*ClusterSlotResult)(&filteredSlots), nil
	}

	return nil, fmt.Errorf("Failed to get `cluster slots` on node ip:%s", r.GetAddr())
}

func (r *Redis) GetAddr() string {
	return r.cli.Options().Addr
}

func (r *Redis) IsMaster() bool {
	res, err := r.cli.Do(context.Background(), command.ROLE).Result()
	if err != nil {
		logger.Errorf("Failed to get 'role' from node %s. Err=%s", r.GetAddr(), err)
		return false
	}

	resStr, ok := res.([]interface{})
	if !ok {
		logger.Errorf("Failed to get 'role' from node %s. Reason= result cannot be parsed to []interface{}", r.GetAddr())
		return false
	}

	if len(resStr) > 0 {
		return resStr[0].(string) == "master"
	}

	return false
}

func (r *Redis) SetSlot(slot string, state state.State, nodeId string) error {
	cmd := append(command.CLUSTER_SETSLOT, slot, string(state), nodeId)

	res, err := r.cli.Do(context.Background(), cmd...).Result()
	if err != nil {
		logger.Errorf("Failed to 'cluster setslot %s %s %s' node. Err=%s", slot, state, nodeId, err)
		return err
	}

	resStr, ok := res.(string)
	if ok && resStr == "OK" {
		return nil
	}

	logger.Infof("Failed to parse the SetSlot result command!")
	return err
}

func (r *Redis) GetKeysInSlot(slot string, limitCount int) []string {
	cmd := append(command.CLUSTER_GETKEYSINSLOT, slot, limitCount)

	res, err := r.cli.Do(context.Background(), cmd...).Result()
	if err != nil {
		logger.Errorf("Failed to 'cluster getkeysinslot %s %d' node. Err=%s", slot, limitCount, err)
		return []string{}
	}

	result, ok := res.([]interface{})
	if !ok {
		logger.Error("Failed to convert result GetKeysInSlot to []interface{}")
		return []string{}
	}

	resultStr := make([]string, len(result))
	for i, v := range result {
		resultStr[i] = fmt.Sprint(v)
	}

	return resultStr
}

func (r *Redis) Reshard(slots []string, destCli IRedis, maxConcurrent int) error {
	if !r.IsMaster() || !destCli.IsMaster() {
		return errors.New("Reshard failed, the both nodes should be master")
	}

	if len(slots) > 1 {
		slotStartInt := parser.ParseInt(slots[0])
		slotEndInt := parser.ParseInt(slots[1])

		// Default max goroutine: 5
		// In other word, Migrate n slots (default: 5) at a time
		var (
			semaphore = make(chan struct{}, maxConcurrent)
			wg        sync.WaitGroup
		)

		for slot := slotStartInt; slot <= slotEndInt; slot++ {
			wg.Add(1)
			go func() {
				defer wg.Done()

				// Add 1 to semaphore
				// Got blocked whenever buffer channel contains 5 elements
				semaphore <- struct{}{}

				// Reduce semaphore when the process finished
				defer func() { <-semaphore }()

				// Perform reshard per 1 slot
				logger.Infof("Performing reshard slot %d from %s to %s", slot, r.GetAddr(), destCli.GetAddr())
				r.processReshard(strconv.Itoa(slot), destCli)
			}()
		}

		wg.Wait()
	} else { // Migrate 1 slot only. non-zero length
		return r.processReshard(slots[0], destCli)
	}

	return nil
}

func (r *Redis) processReshard(slot string, destCli IRedis) error {
	sourceId, err := r.GetClusterMyId()
	if err != nil {
		logger.Errorf("[processReshard] Failed to get source node id. Err=%s", err)
		return err
	}

	destId, err := destCli.GetClusterMyId()
	if err != nil {
		logger.Errorf("[processReshard] Failed to get source node id. Err=%s", err)
		return err
	}

	destAddr := destCli.GetAddr()
	destHost, destPort, err := net.SplitHostPort(destAddr)
	if err != nil {
		logger.Errorf("[processReshard] Failed to split host-port of destination node id. Err=%s", err)
		return err
	}

	// Set SLOT to IMPORTING state on destination node
	err = destCli.SetSlot(slot, state.IMPORTING, sourceId)
	if err != nil {
		logger.Errorf(
			"[processReshard] Failed to SetSlot to IMPORTING on destination node ip:%s. Err=%s",
			destAddr,
			err,
		)
		return err
	}

	// Set SLOT to MIGRATING state on source node
	err = r.SetSlot(slot, state.MIGRATING, destId)
	if err != nil {
		logger.Errorf(
			"[processReshard] Failed to SetSlot to MIGRATING on source node ip:%s. Err=%s",
			r.cli.Options().Addr,
			err,
		)
		return err
	}

	// Note: Unable to do parallel using ClusterGetKeysInSlot because the command is not cursor based.
	// Migrate keys still in serial manner
	keysBatch, errs := []string{}, []error{}
	for {
		keys := r.GetKeysInSlot(slot, constant.DEFAULT_KEYS_COUNT)
		if len(keys) == 0 {
			break
		}

		for _, key := range keys {
			keysBatch = append(keysBatch, key)

			// Once we accumulate a full batch, migrate the keys
			if len(keysBatch) == constant.DEFAULT_MAX_MIGRATE_KEY_COUNT {
				err = r.migrateKey(destHost, destPort, slot, keysBatch)
				if err != nil {
					errs = append(errs, err)
				}
				keysBatch = nil
			}
		}

		// Migrate remaining keys
		if len(keysBatch) > 0 {
			err = r.migrateKey(destHost, destPort, slot, keysBatch)
			if err != nil {
				errs = append(errs, err)
			}
			keysBatch = nil
		}
	}

	// If there is any error happened during the migration
	// we should not set slot ownership to destination node.
	if len(errs) > 0 {
		for _, err := range errs {
			logger.Error(err.Error())
		}

		// Revert the slot state
		for {
			logger.Infof("Reshard failed. Reverting slot %s ownership to source node: %s", slot, r.GetAddr())

			err, err2 := r.SetSlot(slot, state.NODE, sourceId), destCli.SetSlot(slot, state.NODE, sourceId)
			if err != nil || err2 != nil {
				time.Sleep(2 * time.Second)
				continue
			}

			break
		}
		os.Exit(-1)
	}

	// After MIGRATE process finished,
	// SetSlot <slot> NODE <dest_id> on both source and destination node
	err = r.SetSlot(slot, state.NODE, destId)
	if err != nil {
		logger.Errorf(
			"[processReshard] Failed to SetSlot to NODE on source node ip:%s. Err=%s",
			r.cli.Options().Addr,
			err,
		)
		return err
	}

	err = destCli.SetSlot(slot, state.NODE, destId)
	if err != nil {
		logger.Errorf(
			"[processReshard] Failed to SetSlot to NODE on destination node ip:%s. Err=%s",
			destAddr,
			err,
		)
		return err
	}

	logger.Infof("Finished reshard slot %s from %s to %s", slot, r.GetAddr(), destCli.GetAddr())

	return nil
}

func (r *Redis) migrateKey(destHost, destPort, slot string, keys []string) error {
	sourceAddr := r.GetAddr()
	cmd := append(command.MIGRATE, destHost, destPort, "", 0, 5*time.Second, command.REPLACE, command.AUTH, r.cli.Options().Password, command.KEYS)

	for _, key := range keys {
		cmd = append(cmd, key)
	}

	for attempt := 1; attempt <= constant.DEFAULT_MAX_RETRY_MIGRATE_KEY; attempt++ {
		statusCmd := r.cli.Do(context.Background(), cmd...)
		res, err := statusCmd.Result()

		if err != nil {
			logger.Errorf(
				"Failed to MIGRATE keys: <%v> (slot: %s) to destination node %s:6379. Attempt: %d. Err=%s",
				keys,
				slot,
				destHost,
				attempt,
				err,
			)

			// Retry on transient errors
			time.Sleep(time.Duration(attempt) * time.Second)
			continue
		}

		switch res {
		case OK:
			{
				logger.Infof("Successfully migrate keys: <%v> (slot: %s) from %s to %s:6379", keys, slot, sourceAddr, destHost)
				return nil
			}
		case NOKEY:
			{
				logger.Errorf("Got NOKEY error, ignore the key...")
				return nil
			}
		case BUSYKEY:
			{
				logger.Errorf("Got BUSYKEY %v (slot: %s) error, retrying...", keys, slot)
			}
		default:
			logger.Errorf("Unexpected response for keys: %v. resp: %s", keys, res)
		}

		// Exponential backoff for retries
		time.Sleep(time.Duration(attempt) * time.Second)
	}

	return fmt.Errorf("Failed to MIGRATE key: <%v> (slot: %s) to destination node %s", keys, slot, destHost)
}
