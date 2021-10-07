package log

type Level int

const (
	ERROR Level = 0 + 10*iota
	WARNING
	INFO
	DEBUG
)
