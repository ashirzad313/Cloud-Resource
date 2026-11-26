-- Milestone 3, Step 1: the first version of our VM resource table.
-- A migration is a versioned change to the structure of a database.
CREATE TABLE virtual_machines (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    cpu_cores INTEGER NOT NULL,
    memory_mb INTEGER NOT NULL,
    region VARCHAR(50) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'stopped',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT vm_name_not_blank CHECK (length(btrim(name)) > 0),
    CONSTRAINT vm_cpu_cores_valid CHECK (cpu_cores BETWEEN 1 AND 64),
    CONSTRAINT vm_memory_valid CHECK (
        memory_mb BETWEEN 512 AND 262144
        AND memory_mb % 512 = 0
    ),
    CONSTRAINT vm_region_not_blank CHECK (length(btrim(region)) > 0),
    CONSTRAINT vm_status_valid CHECK (
        status IN ('stopped', 'starting', 'running', 'stopping', 'error')
    )
);
