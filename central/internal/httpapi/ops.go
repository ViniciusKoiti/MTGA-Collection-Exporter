package httpapi

import "context"

// OpsStatus is the operations-plane snapshot: readiness facts only,
// never raw configuration or credentials.
type OpsStatus struct {
	SchemaReady   bool  `json:"schema_ready"`
	QueueDepth    int64 `json:"queue_depth"`
	OldestJobAge  int64 `json:"oldest_job_age_seconds"`
	PoolInUse     int32 `json:"pool_in_use"`
	PoolMax       int32 `json:"pool_max"`
	SigningLoaded bool  `json:"signing_loaded"`
}

// OpsJob is the read-only view of one durable job.
type OpsJob struct {
	ID       string `json:"id"`
	Kind     string `json:"kind"`
	Status   string `json:"status"`
	Attempts int    `json:"attempts"`
	Owner    string `json:"owner,omitempty"`
}

// OpsDeletion is one deletion request and its completion evidence.
type OpsDeletion struct {
	InstallationID string `json:"installation_id"`
	Completed      bool   `json:"completed"`
	Proof          string `json:"proof,omitempty"`
}

// OpsAudit is one append-only audit entry.
type OpsAudit struct {
	Actor  string `json:"actor"`
	Action string `json:"action"`
	At     string `json:"at"`
}

// OpsCatalog describes the currently published catalog snapshot.
type OpsCatalog struct {
	SnapshotID  string `json:"snapshot_id"`
	PublishedAt string `json:"published_at"`
	CardCount   int    `json:"card_count"`
}

// OpsSource is the read-only port behind the operations plane: no
// generic SQL, no mutations — each method answers one fixed question.
type OpsSource interface {
	Status(ctx context.Context) (OpsStatus, error)
	Job(ctx context.Context, id string) (OpsJob, error)
	Deletions(ctx context.Context, limit int) ([]OpsDeletion, error)
	Audit(ctx context.Context, limit int) ([]OpsAudit, error)
	Catalog(ctx context.Context) (OpsCatalog, error)
}
