package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/ashirzad313/Cloud-Resource/internal/vm"
	"github.com/jackc/pgx/v5"
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

// Create inserts a VM and returns the complete stored record, including defaults.
func (r *VMRepository) Create(ctx context.Context, input vm.CreateInput) (vm.VirtualMachine, error) {
	const query = `
		INSERT INTO virtual_machines (name, cpu_cores, memory_mb, region)
		VALUES ($1, $2, $3, $4)
		RETURNING id::text, name, cpu_cores, memory_mb, region, status,
		          created_at, updated_at
	`

	var machine vm.VirtualMachine
	err := r.pool.QueryRow(ctx, query,
		input.Name, input.CPUCores, input.MemoryMB, input.Region,
	).Scan(
		&machine.ID,
		&machine.Name,
		&machine.CPUCores,
		&machine.MemoryMB,
		&machine.Region,
		&machine.Status,
		&machine.CreatedAt,
		&machine.UpdatedAt,
	)
	if err != nil {
		return vm.VirtualMachine{}, fmt.Errorf("insert virtual machine: %w", err)
	}

	return machine, nil
}

// Get retrieves a single VM by its primary key.
func (r *VMRepository) Get(ctx context.Context, id string) (vm.VirtualMachine, error) {
	const query = `
		SELECT id::text, name, cpu_cores, memory_mb, region, status,
		       created_at, updated_at
		FROM virtual_machines
		WHERE id = $1::text::uuid
	`

	var machine vm.VirtualMachine
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&machine.ID, &machine.Name, &machine.CPUCores, &machine.MemoryMB,
		&machine.Region, &machine.Status, &machine.CreatedAt, &machine.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return vm.VirtualMachine{}, vm.ErrNotFound
	}
	if err != nil {
		return vm.VirtualMachine{}, fmt.Errorf("get virtual machine: %w", err)
	}
	return machine, nil
}

// Delete removes only a stopped VM. The row lock keeps the status check
// and deletion atomic with respect to concurrent updates and deletes.
func (r *VMRepository) Delete(ctx context.Context, id string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin VM deletion: %w", err)
	}
	defer tx.Rollback(ctx)

	var status string
	err = tx.QueryRow(ctx,
		"SELECT status FROM virtual_machines WHERE id = $1::text::uuid FOR UPDATE",
		id,
	).Scan(&status)
	if errors.Is(err, pgx.ErrNoRows) {
		return vm.ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("lock virtual machine: %w", err)
	}
	if status != "stopped" {
		return vm.ErrConflict
	}

	if _, err := tx.Exec(ctx,
		"DELETE FROM virtual_machines WHERE id = $1::text::uuid", id,
	); err != nil {
		return fmt.Errorf("delete virtual machine: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit VM deletion: %w", err)
	}
	return nil
}
