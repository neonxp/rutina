package rutina

type Options int

const (
	ShutdownIfFail  Options = iota // Shutdown all routines if fail
	RestartIfFail                  // Restart this routine if fail
	DoNothingIfFail                // Do nothing on fail
	ShutdownIfDone                 // Shutdown all routines if this routine done without errors
	RestartIfDone                  // Restart if this routine done without errors
	DoNothingIfDone                // Do nothing if this routine done without errors
)
