package postgres

import (
	"context"
	"fmt"

	"github.com/ashirzad313/Cloud-Resource/internal/vm"
	"github.com/jackc/pgx/v5/pgxpool"
)

// VMRepository handles persistence of virtual machine records in PostgreSQL.
type VMRepository struct {
	pool *pgxpool.Pool
}

func NewVMRepository(pool *pgxpool.Pool) *VMRepository {
	return &VMRepository{pool: pool}
}

// List retrieves all stored VMs. Pagination is added with the REST API later.
func (r *VMRepository) List(ctx context.Context) ([]vm.VirtualMachine, error) {
	const query = `
		SELECT id::text, name, cpu_cores, memory_mb, region, status,
		       created_at, updated_at
		FROM virtual_machines
		ORDER BY created_at, id
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query virtual machines: %w", err)
	}
	defer rows.Close()

	machines := make([]vm.VirtualMachine, 0)
	for rows.Next() {
		var machine vm.VirtualMachine
		if err := rows.Scan(
			&machine.ID,
			&machine.Name,
			&machine.CPUCores,
			&machine.MemoryMB,
			&machine.Region,
			&machine.Status,
			&machine.CreatedAt,
			&machine.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan virtual machine: %w", err)
		}
		machines = append(machines, machine)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate virtual machines: %w", err)
	}
	return machines, nil
}
