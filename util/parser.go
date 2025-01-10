package util

import (
	"regexp"
	"strconv"
	"strings"

	"ivanj26/sonic/constant"
	"ivanj26/sonic/util/logger"
)

func ParseInt(s string) int {
	num, err := strconv.Atoi(s)
	if err != nil {
		logger.Fatalf("Failed to parse string to int")
	}

	return num
}

func ParseSlot(slotStr string) (slots []string) {
	rangeRegex := regexp.MustCompile(constant.REGEX_SLOT_RANGE)
	exactRegex := regexp.MustCompile(constant.REGEX_SLOT_EXACT)

	if rangeRegex.MatchString(slotStr) {
		slots = strings.Split(slotStr, ",")
	}

	if exactRegex.MatchString(slotStr) {
		slots = exactRegex.FindAllString(slotStr, 1)
	}

	if len(slots) == 0 {
		logger.Fatal("Unknown slot format. Please refer to documentation!")
	}

	return
}
