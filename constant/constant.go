package constant

const (
	DEFAULT_MAX_RETRY_MIGRATE_KEY = 3
	DEFAULT_MAX_MIGRATE_KEY_COUNT = 5
	DEFAULT_KEYS_COUNT            = 1000
	DEFAULT_LOG_FILENAME          = "sonic.log"

	INFO_LOG_FILENAME_FMT  = "%s.info.log"
	ERROR_LOG_FILENAME_FMT = "%s.error.log"

	REGEX_SLOT_RANGE = `\d+,\d+`
	REGEX_SLOT_EXACT = `^\d+$`
)
