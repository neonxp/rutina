//go:generate stringer -type=Event
package rutina

// Event represents lifecycle events
type Event int

const (
	EventRoutineStart Event = iota
	EventRoutineStop
	EventRoutineComplete
	EventRoutineError
	EventAppStop
	EventAppComplete
	EventAppError
)

// Hook is function that calls when event fired
// Params:
// ev Event - fired event
// r *Rutina - pointer to rutina
// rid int - ID of routine if present, 0 - otherwise
type Hook func(ev Event, r *Rutina, rid int) error
