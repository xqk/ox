package elect

import (
	"sync"

	"github.com/xqk/ox/pkg/component"
)

var _ component.Manager = &electorComponent{}

type electorComponent struct {
	components    []component.Component
	leaderElector LeaderElector
}

func NewComponent() *electorComponent {
	return &electorComponent{
		components: make([]component.Component, 0),
	}
}

func (e *electorComponent) Start(stop <-chan struct{}) error {
	errCh := make(chan error)
	e.startNonLeaderComponents(stop, errCh)
	e.startLeaderComponents(stop, errCh)

	select {
	case <-stop:
		return nil
	case err := <-errCh:
		return err
	}
}

func (e *electorComponent) AddComponent(components ...component.Component) error {
	e.components = append(e.components, components...)
	return nil
}

func (e *electorComponent) ShouldBeLeader() bool {
	return false
}

func (e *electorComponent) startNonLeaderComponents(stop <-chan struct{}, errCh chan error) {
	for _, item := range e.components {
		if !item.ShouldBeLeader() {
			go func(c component.Component) {
				if err := c.Start(stop); err != nil {
					errCh <- err
				}
			}(item)
		}
	}
}

func (e *electorComponent) startLeaderComponents(stop <-chan struct{}, errCh chan error) {
	var mutex sync.Mutex
	stopCh := make(chan struct{})
	closeCh := func() {
		mutex.Lock()
		defer mutex.Unlock()
		select {
		case <-stopCh:
		default:
			close(stopCh)
		}
	}

	e.leaderElector.AddCallbacks(
		func(phase CallbackPhase) {
			if phase != CallbackPhasePostStarted {
				return
			}
			mutex.Lock()
			defer mutex.Unlock()
			stopCh = make(chan struct{})
			for _, item := range e.components {
				if item.ShouldBeLeader() {
					go func(c component.Component) {
						if err := c.Start(stopCh); err != nil {
							errCh <- err
						}
					}(item)
				}
			}
		},
		func(phase CallbackPhase) {
			if phase != CallbackPhasePostStopped {
				return
			}
			closeCh()
		},
	)

	go e.leaderElector.Start(stop)
	go func() {
		<-stop
		closeCh()
	}()
}
