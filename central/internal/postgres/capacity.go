package postgres

import "fmt"

// Capacity is the reserved database connection budget (design decision
// 9): a hard total below PostgreSQL's max_connections, split across
// process classes. Saturation rejects or delays work — it never grows
// the budget.
type Capacity struct {
	Total              int // reserved connections, below max_connections
	MigrationsReserve  int // held for the dedicated migrator
	OperationsReserve  int // held for operations access
	WorkerConns        int // total for worker processes
	APIConnsPerReplica int // pool ceiling of one API replica
}

// Validate rejects budgets that cannot host even one API replica.
func (c Capacity) Validate() error {
	for name, value := range map[string]int{
		"total": c.Total, "migrations_reserve": c.MigrationsReserve,
		"operations_reserve": c.OperationsReserve, "worker_conns": c.WorkerConns,
		"api_conns_per_replica": c.APIConnsPerReplica,
	} {
		if value < 1 {
			return fmt.Errorf("postgres: capacity %s must be positive", name)
		}
	}
	if c.MaxAPIReplicas() < 1 {
		return fmt.Errorf("postgres: capacity leaves no room for an API replica")
	}
	return nil
}

// apiShare is what remains for API replicas after every reservation.
func (c Capacity) apiShare() int {
	return c.Total - c.MigrationsReserve - c.OperationsReserve - c.WorkerConns
}

// MaxAPIReplicas derives the replica ceiling from the reserved capacity:
// scaling beyond it would overdraw the database, so deployment tooling
// must treat this number as a hard limit.
func (c Capacity) MaxAPIReplicas() int {
	if c.APIConnsPerReplica < 1 {
		return 0
	}
	return c.apiShare() / c.APIConnsPerReplica
}
