package vm

import "time"

// VirtualMachine is the application's representation of one VM record.
// It does not depend on PostgreSQL or HTTP.
type VirtualMachine struct {
	ID        string
	Name      string
	CPUCores  int32
	MemoryMB  int32
	Region    string
	Status    string
	CreatedAt time.Time
	UpdatedAt time.Time
}
