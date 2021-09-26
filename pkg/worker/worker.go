package worker

// Worker could scheduled by ox or customized scheduler
type Worker interface {
	Run() error
	Stop() error
}
