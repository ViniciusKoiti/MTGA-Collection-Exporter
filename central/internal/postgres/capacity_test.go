package postgres

import "testing"

func TestCapacityDerivesReplicaCeiling(t *testing.T) {
	capacity := Capacity{Total: 100, MigrationsReserve: 2, OperationsReserve: 3,
		WorkerConns: 15, APIConnsPerReplica: 10}
	if err := capacity.Validate(); err != nil {
		t.Fatalf("valid capacity rejected: %v", err)
	}
	if got := capacity.MaxAPIReplicas(); got != 8 { // (100-2-3-15)/10
		t.Fatalf("replica ceiling: expected 8, got %d", got)
	}
}

func TestCapacityRejectsOverdraftAndZeros(t *testing.T) {
	overdrawn := Capacity{Total: 20, MigrationsReserve: 5, OperationsReserve: 5,
		WorkerConns: 8, APIConnsPerReplica: 10} // share=2 < one replica
	if err := overdrawn.Validate(); err == nil {
		t.Fatal("capacity without room for one API replica must fail")
	}
	if overdrawn.MaxAPIReplicas() != 0 {
		t.Fatalf("overdrawn ceiling must be 0, got %d", overdrawn.MaxAPIReplicas())
	}
	zeroed := Capacity{Total: 100, APIConnsPerReplica: 10}
	if err := zeroed.Validate(); err == nil {
		t.Fatal("zero reserves must fail validation")
	}
}
