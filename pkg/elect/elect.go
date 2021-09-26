package elect

type CallbackPhase int

const (
	CallbackPhasePostStarted CallbackPhase = 1
	CallbackPhasePostStopped CallbackPhase = 2
)

type LeaderElectCallback func(CallbackPhase)

type LeaderElector interface {
	Start(stop <-chan struct{})
	IsLeader() bool
	AddCallbacks(...LeaderElectCallback)
}
