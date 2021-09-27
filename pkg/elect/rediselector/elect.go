package rediselector

import "github.com/xqk/ox/pkg/elect"

var _ elect.LeaderElector = &redisLeaderElector{}

type redisLeaderElector struct {
	callbacks []elect.LeaderElectCallback
}

func New() *redisLeaderElector {
	return &redisLeaderElector{
		callbacks: make([]elect.LeaderElectCallback, 0),
	}
}

func (r *redisLeaderElector) IsLeader() bool {
	return false
}

func (r *redisLeaderElector) AddCallbacks(callbacks ...elect.LeaderElectCallback) {

}

func (r *redisLeaderElector) Start(stop <-chan struct{}) {
}
