package agents

// StatusCallback is a function that agents call to report their current status.
// The status string should be a concise description of what the agent is doing.
type StatusCallback func(status string)
