package memelector

import "ox/pkg/elect"

type noopLeaderElector struct {
	alwaysLeader bool
	callbacks    []elect.LeaderElectCallback
}

var _ elect.LeaderElector = &noopLeaderElector{}

func NewAlwaysLeaderElector() elect.LeaderElector {
	return &noopLeaderElector{
		alwaysLeader: true,
	}
}

func NewNeverLeaderElector() elect.LeaderElector {
	return &noopLeaderElector{
		alwaysLeader: false,
	}
}

func (n *noopLeaderElector) AddCallbacks(callbacks ...elect.LeaderElectCallback) {
	n.callbacks = append(n.callbacks, callbacks...)
}

func (n *noopLeaderElector) IsLeader() bool {
	return n.alwaysLeader
}

func (n *noopLeaderElector) Start(stop <-chan struct{}) {
	if n.alwaysLeader {
		for _, callback := range n.callbacks {
			callback(elect.CallbackPhasePostStarted)
		}
	}
}
