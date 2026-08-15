package syncstate

import "testing"

func TestLifecycleWalkAndInvalidEventsKeepState(t *testing.T) {
	machine := New()
	walk := []struct {
		event Event
		want  State
	}{
		{EventConfigure, Ready}, {EventWatch, Watching},
		{EventSyncStarted, Syncing}, {EventSyncFailed, Error},
		{EventSyncStarted, Syncing}, {EventSyncSucceeded, Fresh},
		{EventAged, Stale}, {EventSyncStarted, Syncing},
		{EventSyncSucceeded, Fresh},
	}
	for i, step := range walk {
		got, err := machine.Apply(step.event)
		if err != nil || got != step.want {
			t.Fatalf("step %d (%s): got %s (%v), want %s", i, step.event, got, err, step.want)
		}
	}
	if _, err := machine.Apply(EventSyncSucceeded); err == nil {
		t.Fatal("sync_succeeded outside syncing must be invalid")
	}
	if machine.Current() != Fresh {
		t.Fatalf("invalid event must not change state: %s", machine.Current())
	}
	if _, err := New().Apply(EventSyncStarted); err == nil {
		t.Fatal("not_configured cannot start a sync")
	}
}
