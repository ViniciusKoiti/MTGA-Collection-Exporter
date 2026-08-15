// Package syncstate implements the collection sync state machine
// (OpenSpec introduce-agentic-go-companion, task 3.3): the compiled
// transition table is the only way the state advances — UI and jobs
// observe it, they never invent states.
package syncstate

import "fmt"

// State is the sync lifecycle state shown to the user.
type State string

// Lifecycle states.
const (
	NotConfigured State = "not_configured"
	Ready         State = "ready"
	Watching      State = "watching"
	Syncing       State = "syncing"
	Fresh         State = "fresh"
	Stale         State = "stale"
	Error         State = "error"
)

// Event is a typed lifecycle event.
type Event string

// Lifecycle events.
const (
	EventConfigure     Event = "configure"      // a source was configured
	EventWatch         Event = "watch"          // log watching enabled
	EventSyncStarted   Event = "sync_started"   // a sync run began
	EventSyncSucceeded Event = "sync_succeeded" // snapshot committed
	EventSyncFailed    Event = "sync_failed"    // sync run failed
	EventAged          Event = "aged"           // freshness window elapsed
)

// transitions is the compiled table; anything absent is invalid.
var transitions = map[State]map[Event]State{
	NotConfigured: {EventConfigure: Ready},
	Ready:         {EventWatch: Watching, EventSyncStarted: Syncing},
	Watching:      {EventSyncStarted: Syncing},
	Syncing:       {EventSyncSucceeded: Fresh, EventSyncFailed: Error},
	Fresh:         {EventAged: Stale, EventSyncStarted: Syncing, EventWatch: Watching},
	Stale:         {EventSyncStarted: Syncing, EventWatch: Watching},
	Error:         {EventSyncStarted: Syncing, EventConfigure: Ready},
}

// Machine holds the current state; the zero value is not_configured.
type Machine struct {
	current State
}

// New starts the machine at not_configured.
func New() *Machine { return &Machine{current: NotConfigured} }

// Current returns the state without advancing it.
func (m *Machine) Current() State { return m.current }

// Apply advances the machine or fails with a stable error; the state
// never changes on an invalid event.
func (m *Machine) Apply(event Event) (State, error) {
	next, ok := transitions[m.current][event]
	if !ok {
		return m.current, fmt.Errorf(
			"syncstate: event %s is invalid in state %s", event, m.current)
	}
	m.current = next
	return next, nil
}
