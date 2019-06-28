package rutina

// Options sets custom run policies
type Options int

const (
	ShutdownIfError  Options = iota // Shutdown all routines if fail
	RestartIfError                  // Restart this routine if fail
	DoNothingIfError                // Do nothing on fail
	ShutdownIfDone                  // Shutdown all routines if this routine done without errors
	RestartIfDone                   // Restart if this routine done without errors
	DoNothingIfDone                 // Do nothing if this routine done without errors
)
