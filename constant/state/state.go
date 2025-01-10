package state

type State string

const (
	IMPORTING State = "IMPORTING"
	MIGRATING State = "MIGRATING"
	NODE      State = "NODE"
)
