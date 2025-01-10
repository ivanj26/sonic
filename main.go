package main

import (
	"flag"
	"fmt"
	"log"

	redis_lib "ivanj26/sonic/lib"
	"ivanj26/sonic/util"
	"ivanj26/sonic/util/logger"

	"github.com/go-redis/redis/v8"
)

func main() {
	var sourceIp string
	flag.StringVar(&sourceIp, "s", "", "Redis source master ip address")

	var destIp string
	flag.StringVar(&destIp, "d", "", "Redis dest master ip address")

	var password string
	flag.StringVar(&password, "a", "", "Redis password")

	var slotRaw string
	flag.StringVar(&slotRaw, "n", "", "Slots. Could be range 6500,6600 or single slot 6600")

	var maxConcurrent int
	flag.IntVar(&maxConcurrent, "p", 5, "Number of parallel slot migration (equivalent to number of go routine)")

	flag.Parse()

	if util.IsEmpty(sourceIp) ||
		util.IsEmpty(destIp) ||
		util.IsEmpty(password) ||
		util.IsEmpty(slotRaw) {
		logger.Fatal("Please fill the arguments to execute the reshard")
	}

	// Initialize the logger
	err := logger.Initialize("sonic.log")
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}

	// Parse raw slots from user input
	slots := util.ParseSlot(slotRaw)

	// Init source redis client
	sourceCli := redis_lib.NewRedisClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:6379", sourceIp),
		Password: password,
	})
	defer sourceCli.Close()

	// Init destination redis client
	destCli := redis_lib.NewRedisClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:6379", destIp),
		Password: password,
	})
	defer destCli.Close()

	// Perform resharding
	err = sourceCli.Reshard(slots, destCli, maxConcurrent)
	if err != nil {
		logger.Fatalf("Error when performing reshard operation. Err=%s", err)
	}
}
