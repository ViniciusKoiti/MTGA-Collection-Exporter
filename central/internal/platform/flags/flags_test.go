package flags

import "testing"

func TestFeaturesShipDarkAndNeverEnableByAccident(t *testing.T) {
	dark := FromEnv(func(string) string { return "" })
	if dark.CatalogPublication.Enabled() || dark.TelemetryEnrollment.Enabled() {
		t.Fatal("features must ship dark")
	}
	sloppy := FromEnv(func(key string) string { return "TRUE" })
	if sloppy.CatalogPublication.Enabled() {
		t.Fatal("only the exact value \"true\" may enable a feature")
	}
	lit := FromEnv(func(key string) string {
		if key == "CENTRAL_FLAG_CATALOG_PUBLICATION" {
			return "true"
		}
		return ""
	})
	if !lit.CatalogPublication.Enabled() || lit.TelemetryEnrollment.Enabled() {
		t.Fatal("flags must read independently from the environment")
	}
}

// TestRollbackIsIndependentAndImmediate: disabling telemetry never
// touches catalog publication, and vice versa — the staged-rollout
// contract of task 8.7.
func TestRollbackIsIndependentAndImmediate(t *testing.T) {
	set := Set{CatalogPublication: New("catalog_publication", true),
		TelemetryEnrollment: New("telemetry_enrollment", true)}
	set.TelemetryEnrollment.Disable()
	if !set.CatalogPublication.Enabled() {
		t.Fatal("rolling back telemetry must not touch catalog publication")
	}
	if set.TelemetryEnrollment.Enabled() {
		t.Fatal("the rollback must be immediate")
	}
	set.TelemetryEnrollment.Enable()
	set.CatalogPublication.Disable()
	if !set.TelemetryEnrollment.Enabled() || set.CatalogPublication.Enabled() {
		t.Fatal("each feature must flip alone")
	}
}
