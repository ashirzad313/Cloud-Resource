package vm

import "time"

// VirtualMachine represents a VM resource independent of the database driver.
type VirtualMachine struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CPUCores  int32     `json:"cpu_cores"`
	MemoryMB  int32     `json:"memory_mb"`
	Region    string    `json:"region"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
