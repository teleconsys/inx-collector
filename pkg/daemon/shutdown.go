package daemon

const (
	PriorityDisconnectINX = iota // no dependencies
	PriorityStopCollector
	PriorityStopRestAPI
)
