package main

import (
	"flag"
	"fmt"
	"log"

	"ivanj26/sonic/constant"
	redisCli "ivanj26/sonic/lib/redis"
	"ivanj26/sonic/util"
	"ivanj26/sonic/util/logger"
	"ivanj26/sonic/util/parser"

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
	flag.StringVar(&slotRaw, "l", "", "Slots. Could be range 6500,6600 or single slot 6600")

	var nbOfSlot int
	flag.IntVar(&nbOfSlot, "n", 0, "Number of slots. Must be positive integer >0.")

	var maxConcurrent int
	flag.IntVar(&maxConcurrent, "p", 5, "Number of parallel slot migration (equivalent to number of go routine)")

	var logFilePath string
	flag.StringVar(&logFilePath, "log-path", constant.DEFAULT_LOG_FILENAME, "Path to log file")

	var enableInfoLog bool
	flag.BoolVar(&enableInfoLog, "enable-info-log", true, "To disable info log, set as false")

	flag.Parse()

	// Initialize the logger
	err := logger.New().
		SetFilePath(logFilePath).
		SetInfoEnabled(enableInfoLog).
		Initialize()

	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}

	if util.IsEmpty(sourceIp) ||
		util.IsEmpty(destIp) ||
		util.IsEmpty(password) {
		logger.Fatal("Please fill the arguments to execute the reshard")
	}

	// Init source redis client
	sourceCli := redisCli.NewRedisClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:6379", sourceIp),
		Password: password,
	})
	defer sourceCli.Close()

	var totalSlotsRange [][]string
	if util.IsEmpty(slotRaw) {
		if nbOfSlot <= 0 {
			logger.Fatal("Please fill the range of slot or number of slots to migrate!")
		}

		slots, err := sourceCli.GetClusterSlots()
		if slots == nil || err != nil {
			logger.Fatal(err.Error())
		}

		totalSlotsRange = slots.LimitTo(nbOfSlot)
	} else {
		// Parse raw slots from user input
		totalSlotsRange = append(totalSlotsRange, parser.ParseSlot(slotRaw))
	}

	// Init destination redis client
	destCli := redisCli.NewRedisClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:6379", destIp),
		Password: password,
	})
	defer destCli.Close()

	// Perform resharding
	for _, slotRange := range totalSlotsRange {
		if len(slotRange) < 2 {
			logger.Errorf("Invalid slot range. Slots: %+v", slotRange)
			continue
		}

		logger.Infof("-----------------------------------------------------------------")
		logger.Infof("Migrating slots [%s, %s] from %s to %s...", slotRange[0], slotRange[1], sourceIp, destIp)
		logger.Infof("-----------------------------------------------------------------")

		err := sourceCli.Reshard(slotRange, destCli, maxConcurrent)
		if err != nil {
			logger.Fatalf("Error when performing reshard operation on slot range %v. Err=%s", slotRange, err)
		}
	}
}
