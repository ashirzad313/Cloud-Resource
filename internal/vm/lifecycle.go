package vm

// Lifecycle states match the CHECK constraint in PostgreSQL.
const (
	StatusStopped  = "stopped"
	StatusStarting = "starting"
	StatusRunning  = "running"
	StatusStopping = "stopping"
	StatusError    = "error"
)
